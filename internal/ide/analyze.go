package ide

import (
	"fmt"
	"time"

	"github.com/Oridjinnn/Rewind/pkg/types"
)

// AnalyzeProject generates productivity insights from IDE activity data.
func AnalyzeProject(r Recorder, projectPath string) (types.ProductivityInsight, error) {
	insight := types.ProductivityInsight{Project: projectPath}

	// Get all activities for this project
	filter := types.IDEActivityFilter{
		ProjectPath: projectPath,
		Limit:       10000,
	}
	activities, err := r.QueryActivity(filter)
	if err != nil {
		return insight, err
	}
	if len(activities) == 0 {
		return insight, fmt.Errorf("no activity found for project: %s", projectPath)
	}

	insight.TotalEvents = int64(len(activities))

	// Calculate active time
	if len(activities) > 1 {
		first, _ := time.Parse(time.RFC3339, activities[len(activities)-1].ExecutedAt)
		last, _ := time.Parse(time.RFC3339, activities[0].ExecutedAt)
		insight.ActiveHours = last.Sub(first).Hours()
	}

	// Aggregate stats
	langMap := make(map[string]int)
	fileMap := make(map[string]int)
	cmdMap := make(map[string]int)
	var buildTotal, buildSuccess int64
	var testTotal, testPass int64
	var aiTotal, aiReject int64

	for _, a := range activities {
		// Language usage
		if a.Language != "" {
			langMap[a.Language]++
		}
		// File edits
		if a.FilePath != "" && (a.ActivityType == "file_save" || a.ActivityType == "file_edit") {
			fileMap[a.FilePath]++
		}
		// Build stats
		if a.ActivityType == "build_end" || a.ActivityType == "build_start" {
			buildTotal++
			if ec, ok := a.Metadata["exit_code"]; ok {
				if code, ok := ec.(float64); ok && code == 0 {
					buildSuccess++
				}
			}
		}
		// Test stats
		if a.ActivityType == "test_run" || a.ActivityType == "test_pass" {
			testTotal++
		}
		if a.ActivityType == "test_pass" {
			testPass++
		}
		// AI usage
		if a.ActivityType == "ai_chat" || a.ActivityType == "ai_completion" {
			aiTotal++
		}
		if a.ActivityType == "ai_reject" {
			aiReject++
		}
		// Commands
		if a.ActivityType == "terminal_cmd" {
			cmdMap[a.ActivityType]++
		}
	}

	// Top languages
	insight.TopLanguages = topNFromMap(langMap, 5)
	// Top files
	insight.TopFiles = topNFromMap(fileMap, 5)
	// Most used commands from shellhistory would be queried separately
	insight.MostUsedCommands = []string{"(use 'rewind history-stats')"}

	// Rates
	if buildTotal > 0 {
		insight.BuildSuccessRate = float64(buildSuccess) / float64(buildTotal) * 100
	}
	if testTotal > 0 {
		insight.TestPassRate = float64(testPass) / float64(testTotal) * 100
	}
	insight.AIUsageCount = aiTotal
	if aiTotal > 0 {
		insight.AIRejectRate = float64(aiReject) / float64(aiTotal) * 100
	}

	// Generate suggestions
	insight.Suggestions = generateSuggestions(insight)

	return insight, nil
}

// PrintInsight displays productivity insight in human-readable format.
func PrintInsight(insight types.ProductivityInsight) {
	fmt.Println("")
	fmt.Println("PRODUCTIVITY INSIGHT")
	fmt.Println("=====================")
	fmt.Printf("Project:            %s\n", insight.Project)
	fmt.Printf("Total events:       %d\n", insight.TotalEvents)
	fmt.Printf("Active hours:       %.1f\n", insight.ActiveHours)
	fmt.Println("")

	if len(insight.TopLanguages) > 0 {
		fmt.Println("Top Languages:")
		for i, lang := range insight.TopLanguages {
			fmt.Printf("  %d. %s\n", i+1, lang)
		}
		fmt.Println("")
	}

	if len(insight.TopFiles) > 0 {
		fmt.Println("Most Edited Files:")
		for i, f := range insight.TopFiles {
			fmt.Printf("  %d. %s\n", i+1, f)
		}
		fmt.Println("")
	}

	if insight.BuildSuccessRate > 0 || insight.TestPassRate > 0 {
		fmt.Println("Quality Metrics:")
		if insight.BuildSuccessRate > 0 {
			fmt.Printf("  Build success:  %.0f%%\n", insight.BuildSuccessRate)
		}
		if insight.TestPassRate > 0 {
			fmt.Printf("  Test pass:      %.0f%%\n", insight.TestPassRate)
		}
		fmt.Println("")
	}

	if insight.AIUsageCount > 0 {
		fmt.Println("AI Assistant:")
		fmt.Printf("  Interactions:   %d\n", insight.AIUsageCount)
		fmt.Printf("  Reject rate:    %.0f%%\n", insight.AIRejectRate)
		fmt.Println("")
	}

	if len(insight.Suggestions) > 0 {
		fmt.Println("Suggestions:")
		for _, s := range insight.Suggestions {
			fmt.Printf("  → %s\n", s)
		}
		fmt.Println("")
	}
}

func topNFromMap(m map[string]int, n int) []string {
	type pair struct {
		key   string
		count int
	}
	var pairs []pair
	for k, v := range m {
		pairs = append(pairs, pair{k, v})
	}
	// Simple sort (for production use sort.Slice)
	for i := 0; i < len(pairs)-1; i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].count > pairs[i].count {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
	var result []string
	for i := 0; i < n && i < len(pairs); i++ {
		result = append(result, pairs[i].key)
	}
	return result
}

func generateSuggestions(insight types.ProductivityInsight) []string {
	var s []string

	if insight.BuildSuccessRate > 0 && insight.BuildSuccessRate < 50 {
		s = append(s, "Build success rate is low. Consider reviewing recent build errors for patterns.")
	}
	if insight.TestPassRate > 0 && insight.TestPassRate < 70 {
		s = append(s, "Test pass rate could improve. Try running tests more frequently during development.")
	}
	if insight.AIRejectRate > 50 {
		s = append(s, "High AI rejection rate. The model may not be matching your coding style—try different prompts.")
	}
	if insight.ActiveHours > 8 {
		s = append(s, "Long coding sessions detected. Consider taking regular breaks for productivity.")
	}
	if insight.TotalEvents > 1000 && insight.AIUsageCount == 0 {
		s = append(s, "No AI usage detected. Try enabling AI completions for faster coding.")
	}

	if len(s) == 0 {
		s = append(s, "Looking good! Keep up the productive workflow.")
	}
	return s
}