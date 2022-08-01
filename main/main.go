package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/agusnavce/ta"
	"github.com/agusnavce/ta/utils"
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

const MIN_FREQ = 50

func new_main(enteries []*utils.Entry) {
	model := ta.NewSpellModel()

	for _, en := range enteries {
		if en.Frequency >= MIN_FREQ {
			model.AddEntry(*en)
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

		segment, _ := model.Segment(text)
		if strings.Contains(segment.String(), " ") {
			fmt.Println("SEGMENT: ", segment)
		}

	}
}

func main() {
	enteries := readFile("data/new_new_word.json")
	fmt.Println(len(enteries))
	fmt.Printf("%T \n---------------------------\n", enteries)
	// original_main()
	new_main(enteries)
}
