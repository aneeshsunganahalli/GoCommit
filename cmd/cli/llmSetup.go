package cmd

import (
	"errors"
	"fmt"

	"github.com/dfanso/commit-msg/cmd/cli/store"
	"github.com/manifoldco/promptui"
)


func SetupLLM() error {

	providers := []string{"OpenAI", "Claude", "Gemini", "Grok", "Ollama"}
	prompt := promptui.Select{
		Label: "Select LLM",
		Items: providers,
	}

	_, model, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("prompt failed")
	}

	var apiKey string
	
	// Skip API key prompt for Ollama (local LLM)
	apiKeyPrompt := promptui.Prompt{
			Label: "Enter API Key",
			Mask: '*',
		}
	
	
		switch model {
		case "Ollama":
			urlPrompt := promptui.Prompt{
			Label: "Enter URL",
			}
			apiKey, err = urlPrompt.Run()
			if err != nil {
			return fmt.Errorf("failed to read Url: %w", err)
			}

		default:
			apiKey, err = apiKeyPrompt.Run()
			if err != nil {
			return fmt.Errorf("failed to read API Key: %w", err)
		}

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
	
	SavedModels, err := store.ListSavedModels()
	if err != nil {
		return err
	}

	if len(SavedModels.LLMProviders) == 0 {
		 return errors.New("no model exists, Please add atleast one model Run: 'commit llm setup'")
		 
	}

	models := []string{}
	options1 := []string{"Set Default", "Change API Key", "Delete"}
	options2 := []string{"Set Default", "Change URL", "Delete"} //different option for local model

	for _, p := range SavedModels.LLMProviders {
		models = append(models, p.LLM)
	}

	prompt := promptui.Select{
		Label: "Select from saved models",
		Items: models,
	}

	_,model,err := prompt.Run()
	if err != nil {
		return err
	}

		prompt = promptui.Select{
		Label: "Select Option",
		Items: options1,
		}

		apiKeyPrompt := promptui.Prompt {
		Label: "Enter API Key",
		}

	
		if model == "Ollama" {
			prompt = promptui.Select{
			Label: "Select Option",
				Items: options2,
			}

			apiKeyPrompt = promptui.Prompt {
				Label: "Enter URL",
			}
		}


	opNo,_,err := prompt.Run()
	if err != nil {
		return err
	}



	switch opNo {
		case 0:
			err := store.ChangeDefault(model)
			if err != nil {
				return err
			}
			fmt.Printf("%s set as default", model)
		case 1:
			apiKey, err := apiKeyPrompt.Run()
			if err !=  nil {
				return err
			}
			err = store.UpdateAPIKey(model, apiKey)
			if err != nil {
				return err
			}
			event := "API Key"
			if model == "Ollama"{event = "URL"} 
			fmt.Printf("%s %s Updated", model,event)
		case 2:
			err := store.DeleteModel(model)
			if err != nil {
				return err
			}
			fmt.Printf("%s model deleted", model)			
	}

	return nil
}
