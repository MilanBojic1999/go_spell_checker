package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	pb "go_spell_checker/proton"
	"go_spell_checker/ta"
	"go_spell_checker/utils"
	"os"
	"regexp"
	"strings"
)

const (
	MIN_FREQ = 50
)

type SpellCorrectorService interface {
	Correction(context.Context, pb.SpellCorrectorInput) (pb.SpellCorrectorOutput, error)
}

type SpellCheckerServer struct {
	pb.SpellCorrectorServiceClient
	Model  *ta.SpellModel
	server bool
	reg    *regexp.Regexp
}

func (server *SpellCheckerServer) Correction(cont context.Context, input *pb.SpellCorrectorInput) (*pb.SpellCorrectorOutput, error) {
	query := input.Query

	query = server.reg.ReplaceAllString(query, "")
	query = strings.ToLower(query)
	result, err := server.Model.LookupCompund(query)
	if err != nil || len(result) == 0 {
		ret := pb.SpellCorrectorOutput{Result: "", IsCorrected: false, Probability: 0.0, Language: "None"}
		return &ret, err
	}
	toReturn := result[0]
	fmt.Println(query, " @@@ ", result)
	if strings.Compare(query, toReturn.Word) == 0 {
		ret := pb.SpellCorrectorOutput{Result: "", IsCorrected: false, Probability: 0.0, Language: "None"}
		return &ret, nil
	}
	return &pb.SpellCorrectorOutput{Result: toReturn.Word, IsCorrected: true, Probability: float32(toReturn.Frequency), Language: "eng"}, nil
}

func readFile(path string) []*utils.Entry {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}

	defer file.Close()

	result := make([]*utils.Entry, 0)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		scanvalue := new(utils.Entry)
		json.Unmarshal([]byte(scanner.Text()), &scanvalue)
		result = append(result, scanvalue)

	}

	return result
}

func getModel() *ta.SpellModel {
	model := ta.NewSpellModel()

	enteries_words := readFile("data/words_jj.json")
	enteries_bigrams := readFile("data/bigrams_jj.json")

	for _, en := range enteries_words {
		if en.Frequency >= MIN_FREQ {
			model.AddEntry(*en)
		}
	}

	for _, en := range enteries_bigrams {
		if en.Frequency >= MIN_FREQ {
			model.AddBigram(*en)
		}
	}

	return model
}

func NewServer() *SpellCheckerServer {
	s := new(SpellCheckerServer)
	s.Model = getModel()
	s.reg, _ = regexp.Compile("[^a-zA-Z0-9]+")

	return s
}
