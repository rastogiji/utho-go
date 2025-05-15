package utho

import (
	"errors"
)

type ApiKeyService service

type ApiKeys struct {
	Status  string   `json:"status,omitempty"`
	Message string   `json:"message,omitempty"`
	API     []ApiKey `json:"api"`
}

type ApiKey struct {
	ID        string `json:"id" faker:"oneof:176300,176302,176303,176304,176305,176306,176507,176508,177492,177499,245495"`
	Name      string `json:"name" faker:"oneof:name,nam122e,name22,testkey,,testq,test-key,PostMan"`
	Write     string `json:"write" faker:"oneof:0,1"`
	CreatedAt string `json:"created_at" faker:"date"`
}

type CreateApiKeyParams struct {
	Name  string `json:"name"`
	Write string `json:"write"`
}

type CreateApiKeyResponse struct {
	Status  string `json:"status"`
	Apikey  string `json:"apikey"`
	Message string `json:"message"`
}

func (s *ApiKeyService) Create(params CreateApiKeyParams) (*CreateApiKeyResponse, error) {
	reqUrl := "api/generate"
	req, _ := s.client.NewRequest("POST", reqUrl, &params)

	var apiKey CreateApiKeyResponse
	_, err := s.client.Do(req, &apiKey)
	if err != nil {
		return nil, err
	}
	if apiKey.Status != "success" && apiKey.Status != "" {
		return nil, errors.New(apiKey.Message)
	}
	return &apiKey, nil
}

func (s *ApiKeyService) List() ([]ApiKey, error) {
	reqUrl := "api"
	req, _ := s.client.NewRequest("GET", reqUrl)

	var apikeys ApiKeys
	_, err := s.client.Do(req, &apikeys)
	if err != nil {
		return nil, err
	}
	if apikeys.Status != "success" && apikeys.Status != "" {
		return nil, errors.New(apikeys.Message)
	}

	return apikeys.API, nil
}

func (s *ApiKeyService) Delete(apiKeyId string) (*DeleteResponse, error) {
	reqUrl := "api/" + apiKeyId + "/delete"
	req, _ := s.client.NewRequest("DELETE", reqUrl)

	var delResponse DeleteResponse
	if _, err := s.client.Do(req, &delResponse); err != nil {
		return nil, err
	}
	if delResponse.Status != "success" && delResponse.Status != "" {
		return nil, errors.New(delResponse.Message)
	}

	return &delResponse, nil
}
