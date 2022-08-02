package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"go_spell_checker/server"
	ta "go_spell_checker/ta"
	utils "go_spell_checker/utils"

	"go_spell_checker/proton"

	"google.golang.org/grpc"
)

func original_main() {
	// Create a new instance of model
	t := ta.NewSpellModel()

	// Add words to the dictionary. Words require a frequency, but can have
	// other arbitrary metadata associated with them
	t.AddEntry(utils.Entry{
		Frequency: 100,
		Word:      "word",
		WordData: utils.WordData{
			"type": "noun",
		},
	})

	t.AddEntry(utils.Entry{
		Frequency: 1,
		Word:      "world",
		WordData: utils.WordData{
			"type": "noun",
		},
	})

	// Lookup a mismodeling, by default the "best" suggestion will be returned
	suggestions, _ := t.Lookup("wortd")
	fmt.Println(suggestions)
	// -> [word]

	suggestion := suggestions[0]

	// Get the frequency from the suggestion
	fmt.Println(suggestion.Frequency)
	// -> 100

	// Get metadata from the suggestion
	fmt.Println(suggestion.WordData["type"])
	// -> noun

	// Get multiple suggestions during lookup
	suggestions, _ = t.Lookup("wortd", ta.SuggestionLevel(ta.ALL))
	fmt.Println(suggestions)
	// -> [word, world]

	// Add multiple entries in one command. You can override the previos config or
	// have a cumulative behaviour
	t.AddEntries(utils.Entries{
		Words: []string{"word", "word"},
		WordsData: utils.WordData{
			"type": "other",
		},
	}, ta.OverrideFrequency(true))

	// Save the dictionary
	t.Save("dict.model")

	// Load the dictionary
	t2, _ := ta.Load("dict.model")

	suggestions, _ = t2.Lookup("wortd", ta.SuggestionLevel(ta.ALL))
	fmt.Println(suggestions)
	// -> [word, world]

	// Create a Dictionary from a file, merges with any dictionary data already loaded.
	t2.CreateDictionary("test.txt")

	entry, err := t2.GetEntry("four")

	if err == nil {
		fmt.Println(entry.Word)
		// -> four
	}

	// Spell supports word segmentation
	t3 := ta.NewSpellModel()

	t3.AddEntry(utils.Entry{Frequency: 1, Word: "near"})
	t3.AddEntry(utils.Entry{Frequency: 1, Word: "the"})
	t3.AddEntry(utils.Entry{Frequency: 1, Word: "fireplace"})

	segmentResult, _ := t3.Segment("nearthefireplace")
	fmt.Println(segmentResult)
	// -> near the fireplace

	// Spell supports multiple dictionaries
	t4 := ta.NewSpellModel()

	t4.AddEntry(utils.Entry{Word: "quindici"}, ta.DictionaryName("italian"))
	suggestions, _ = t4.Lookup("quindici", ta.DictionaryOpts(
		ta.DictionaryName("italian"),
	))
	fmt.Println(suggestions)
	// -> [quindici]
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
		if scanvalue.Frequency < MIN_FREQ {
			continue
		}
		result = append(result, scanvalue)

	}

	return result
}

const (
	MIN_FREQ = 50
	PORT     = "8090"
)

func new_main(enteries_words, enteries_bigrams []*utils.Entry) {
	model := ta.NewSpellModel()

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

	fmt.Print("READY \n*************************\n")

	listen(*model)
	fmt.Println("QUITING....")
}

func listen(model ta.SpellModel) {

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(":> ")

		text, err := reader.ReadString('\n')
		if err != nil {
			// close channel just to inform others
			fmt.Println("Error in read string", err)
		}
		text = strings.Replace(text, "\n", "", -1)

		if len(text) == 0 {
			break
		}
		fmt.Println(text)
		suggestion, _ := model.Lookup(text)
		if len(suggestion) > 0 {
			fmt.Printf("BEST: %s [%d]\n", suggestion, suggestion[0].Distance)
		} else {
			fmt.Println("BEST: ", suggestion)
		}
		suggestion, _ = model.Lookup(text, ta.SuggestionLevel(ta.CLOSEST))
		fmt.Println("CLOSEST: ", suggestion)

		suggestion, _ = model.Lookup(text, ta.SuggestionLevel(ta.ALL))
		if len(suggestion) < 10 {
			fmt.Println("ALL: ", suggestion)
		}

		suggestion, _ = model.LookupCompund(text, ta.SuggestionLevel(ta.ALL))
		fmt.Println("COMPOUND: ", suggestion)

		segment, _ := model.Segment(text)
		fmt.Println("SEGMENT: ", segment)

	}
}

func main_two() {
	enteries_words := readFile("data/words_jj.json")
	enteries_bigrams := readFile("data/bigrams_jj.json")
	fmt.Println(len(enteries_words))
	fmt.Printf("%T \n---------------------------\n", enteries_words)
	fmt.Println(len(enteries_bigrams))
	fmt.Printf("%T \n---------------------------\n", enteries_bigrams)
	// original_main()
	new_main(enteries_words, enteries_bigrams)
}

func main() {
	fmt.Println("STARTING server on port: ", PORT)

	laddr, err := net.ResolveTCPAddr("tcp", "localhost:"+PORT)
	if err != nil {
		fmt.Printf("ERROR with connection: %v\n", err)
		return
	}

	lis, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		fmt.Printf("ERROR with connection: %v\n", err)
		return
	}
	newServerSC := server.NewServer()
	grpcServer := grpc.NewServer()
	proton.RegisterSpellCorrectorServiceServer(grpcServer, newServerSC)

	if err := grpcServer.Serve(lis); err != nil {
		fmt.Printf("failed to serve: %v", err)
	}
}
