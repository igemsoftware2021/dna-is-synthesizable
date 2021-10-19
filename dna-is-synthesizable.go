package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/Open-Science-Global/idtapi"
	"github.com/Open-Science-Global/poly/io/genbank"
	"github.com/spf13/cobra"
)

func main() {
	Execute()
}

type SeqInfo struct {
	Name            string           `json:"name"`
	IsSynthesizable bool             `json:"isSynthesizable"`
	ComplexityScore float64          `json:"complexityScore"`
	Sequence        string           `json:"sequence"`
	Problems        []idtapi.Problem `json:"problems"`
}

var (
	input           string
	output          string
	pattern         string
	idtUsername     string
	idtPassword     string
	idtClientId     string
	idtClientSecret string
	isAlert         bool
)

var TokenURL = "https://www.idtdna.com/Identityserver/connect/token"
var ComplexityScoreURL = "https://www.idtdna.com/api/complexities/screengBlockSequences"
var MaxIDTSequenceLength = 3000

var rootCmd = &cobra.Command{
	Use:   "dna-is-synthesizable",
	Short: "A github action to check if a part is synthesizable.",
	Long:  "A github action to check if a part is synthesizable from a given Genbank file.",
	Run: func(cmd *cobra.Command, args []string) {
		Script(input, output, pattern, isAlert, idtUsername, idtPassword, idtClientId, idtClientSecret)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&input, "input", "i", "", "Directory where all the input genbank files will be read")
	rootCmd.PersistentFlags().StringVarP(&output, "ouput", "o", "", "Directory where all the output genbank files wil be written")
	rootCmd.PersistentFlags().StringVarP(&pattern, "pattern", "r", "", "Regex to filter files in the input directory")
	rootCmd.PersistentFlags().BoolVarP(&isAlert, "alert", "a", false, "Display an error when a non-synthesizable sequence is found")

	rootCmd.PersistentFlags().StringVarP(&idtUsername, "username", "u", "", "IDT account username")
	rootCmd.PersistentFlags().StringVarP(&idtPassword, "password", "p", "", "IDT account password")
	rootCmd.PersistentFlags().StringVarP(&idtClientId, "clientId", "c", "", "IDT API ClientId")
	rootCmd.PersistentFlags().StringVarP(&idtClientSecret, "clientSecret", "s", "", "IDT API ClientSecret")

	rootCmd.MarkFlagRequired("input")
	rootCmd.MarkFlagRequired("ouput")
	rootCmd.MarkFlagRequired("pattern")
	rootCmd.MarkFlagRequired("username")
	rootCmd.MarkFlagRequired("password")
	rootCmd.MarkFlagRequired("clientId")
	rootCmd.MarkFlagRequired("clientSecret")
}

func Script(inputDir string, outputDir string, pattern string, isAlert bool, username string, password string, clientId string, clientSecret string) {
	filesPath := getListFilesByPattern(inputDir, pattern)
	var sequences []idtapi.Sequence
	for _, filePath := range filesPath {
		sequence := genbank.Read(filePath)
		sequences = append(sequences, idtapi.Sequence{sequence.Meta.Locus.Name, strings.ToUpper(sequence.Sequence)})
	}

	infos := sequencesAreSynthesizable(sequences, username, password, clientId, clientSecret)
	writeJsonFile(infos, "synthesizable.json", outputDir)

	if isAlert {
		checkAndAlert(infos)
	}
}

func checkAndAlert(infos []SeqInfo) {
	haveProblems := false
	for _, info := range infos {
		if !info.IsSynthesizable {
			haveProblems = true
			fmt.Printf("The sequence %s can't be synthesized.\nSequence: %s\n\n", info.Name, info.Sequence)
		}
	}

	if haveProblems {
		log.Fatalln("WARNING: fatal error. Check the problems in the output file and try again.")
	}

}

func writeJsonFile(data []SeqInfo, fileName string, outputDir string) {
	outputPath := outputDir + "/"

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		os.Mkdir(outputDir, 0755)
	}

	json, _ := json.Marshal(data)

	filePath := outputPath + fileName

	err := ioutil.WriteFile(filePath, json, 0644)

	if err != nil {
		log.Fatal(err)
	}
}

func getListFilesByPattern(inputDir string, pattern string) []string {
	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		log.Fatal(err)
	}
	var filesPath []string
	for _, f := range files {
		var validFile = regexp.MustCompile(pattern)
		if validFile.MatchString(f.Name()) {
			file := inputDir + "/" + f.Name()
			filesPath = append(filesPath, file)
		}
	}
	return filesPath
}

func sequencesAreSynthesizable(sequences []idtapi.Sequence, username string, password string, clientId string, clientSecret string) []SeqInfo {

	sequencesComplexityScore := idtapi.GetComplexityScore(sequences, username, password, clientId, clientSecret, ComplexityScoreURL, TokenURL)
	var infos []SeqInfo
	for index, sequence := range sequences {
		isSynthesizable := false
		score := calculateComplexityScore(sequencesComplexityScore[index])

		if score < 10.0 && len(sequence.Sequence) <= MaxIDTSequenceLength {
			isSynthesizable = true
		}

		infos = append(infos, SeqInfo{
			sequence.Name,
			isSynthesizable,
			score,
			sequence.Sequence,
			sequencesComplexityScore[index],
		})
	}
	return infos

}

func calculateComplexityScore(problems []idtapi.Problem) float64 {
	score := 0.0
	for _, problem := range problems {
		score += problem.ComplexityScore
	}
	return score
}
