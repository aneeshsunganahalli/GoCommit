package cmd

import (
	"fmt"

	"github.com/dfanso/commit-msg/cmd/cli/store"
	"github.com/manifoldco/promptui"
)


func SetupLLM() error {

	providers := []string{"OpenAI", "Claude", "Gemini", "Grok"}
	prompt := promptui.Select{
		Label: "Select LLM",
		Items: providers,
	}

	_, model, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("prompt failed")
	}

	apiKeyPrompt := promptui.Prompt{
		Label: "Enter API Key",
		
	}

	apiKey, err := apiKeyPrompt.Run()
	if err != nil {
		return fmt.Errorf("invalid API Key")
	}

	LLMConfig := store.LLMProvider{
		LLM:    model,
		APIKey: apiKey,
	}



	err = store.Save(LLMConfig)
	if err != nil {
		return err
	}

	fmt.Println("LLM model added")
	return nil
}

func UpdateLLM() error {
	fmt.Println("Update LLM config")
	return nil
}
