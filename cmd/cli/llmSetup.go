package cmd

import (
	"errors"
	"fmt"

	"github.com/dfanso/commit-msg/cmd/cli/store"
	"github.com/dfanso/commit-msg/pkg/types"
	"github.com/manifoldco/promptui"
)



// SetupLLM walks the user through selecting an LLM provider and storing the
// corresponding API key or endpoint configuration.
func SetupLLM(Store *store.StoreMethods) error {

	providers := types.GetSupportedProviderStrings()
	prompt := promptui.Select{
		Label: "Select LLM",
		Items: providers,
	}

	_, modelStr, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("prompt failed")
	}

	model, valid := types.ParseLLMProvider(modelStr)
	if !valid {
		return fmt.Errorf("invalid LLM provider: %s", modelStr)
	}

	var apiKey string

	// Skip API key prompt for Ollama (local LLM)
	apiKeyPrompt := promptui.Prompt{
		Label: "Enter API Key",
		Mask:  '*',
	}

	switch model {
	case types.ProviderOllama:
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

	err = Store.Save(LLMConfig)
	if err != nil {
		return err
	}

	fmt.Println("LLM model added")
	return nil
}

// UpdateLLM lets the user switch defaults, rotate API keys, or delete stored
// LLM provider configurations.
func UpdateLLM(Store *store.StoreMethods) error {

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
		models = append(models, p.String())
	}

	prompt := promptui.Select{
		Label: "Select from saved models",
		Items: models,
	}

	_, model, err := prompt.Run()
	if err != nil {
		return err
	}

	prompt = promptui.Select{
		Label: "Select Option",
		Items: options1,
	}

	apiKeyPrompt := promptui.Prompt{
		Label: "Enter API Key",
	}

	if model == types.ProviderOllama.String() {
		prompt = promptui.Select{
			Label: "Select Option",
			Items: options2,
		}

		apiKeyPrompt = promptui.Prompt{
			Label: "Enter URL",
		}
	}

	opNo, _, err := prompt.Run()
	if err != nil {
		return err
	}

	switch opNo {
	case 0:
		modelProvider, valid := types.ParseLLMProvider(model)
		if !valid {
			return fmt.Errorf("invalid LLM provider: %s", model)
		}
		err := store.ChangeDefault(modelProvider)
		if err != nil {
			return err
		}
		fmt.Printf("%s set as default", model)
	case 1:
		apiKey, err := apiKeyPrompt.Run()
		if err != nil {
			return err
		}
		modelProvider, valid := types.ParseLLMProvider(model)
		if !valid {
			return fmt.Errorf("invalid LLM provider: %s", model)
		}
		err = Store.UpdateAPIKey(modelProvider, apiKey)
		if err != nil {
			return err
		}
		event := "API Key"
		if model == types.ProviderOllama.String() {
			event = "URL"
		}
		fmt.Printf("%s %s Updated", model, event)
	case 2:
		modelProvider, valid := types.ParseLLMProvider(model)
		if !valid {
			return fmt.Errorf("invalid LLM provider: %s", model)
		}
		err := Store.DeleteModel(modelProvider)
		if err != nil {
			return err
		}
		fmt.Printf("%s model deleted", model)
	}

	return nil
}
