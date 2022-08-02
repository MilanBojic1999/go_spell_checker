package server

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	pb "go_spell_checker/proton"
	"go_spell_checker/ta"
	"go_spell_checker/utils"
	"net"
	"os"
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
}

func (server *SpellCheckerServer) Correction(cont context.Context, input *pb.SpellCorrectorInput) (*pb.SpellCorrectorOutput, error) {
	query := input.Query
	result, err := server.Model.LookupCompund(query)
	if err != nil || len(result) == 0 {
		ret := pb.SpellCorrectorOutput{Result: "", IsCorrected: false, Probability: 0.0, Language: "None"}
		return &ret, err
	}
	toReturn := result[0]
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

func (ser *SpellCheckerServer) Serve(listen net.Listener) error {
	fmt.Println("SERVING")

	if listen == nil {
		fmt.Println("ERORR")
		return errors.New("PORT is null")
	}

	for {
		inMsg, err := listen.Accept()
		if err != nil {
			fmt.Printf("ERror %v\n", err)
		}

		var msgStruct pb.SpellCorrectorInput
		json.NewDecoder(inMsg).Decode(&msgStruct)

		inMsg.Close()
		go func() {
			ser.Correction(nil, &msgStruct)
		}()
	}
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
	return s
}
