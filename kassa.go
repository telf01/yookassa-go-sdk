package yookassa

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	yookassa "yookassa/models"
)

type Kassa struct {
	// Идентификатор вашего магазина.
	ShopID  string
	// Ваш секретный ключ.
	Token   string
	// Режим отображения отладочной информации.
	Verbose bool
	// HTTP клиент для обработки запросов.
	Client  http.Client
	// Адрес, по которому требуется выполнять запросы.
	endpoint string
}

// NewKassa создаёт экземпляр структуры Kassa.
func NewKassa(shopID, token string) *Kassa {
	return &Kassa{
		ShopID:   shopID,
		Token:    token,
		Verbose:  false,
		Client:   http.Client{},
		endpoint: yookassa.APIEndpoint,
	}
}

// Ping отправляет тестовый запрос для проверки соединения.
func (k *Kassa) Ping() (bool, error) {
	resp, err := k.sendGetRequest(yookassa.PaymentsEndpoint, nil)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	return true, nil
}

// SendPaymentConfig отправляет PaymentConfig на сервера ЮКассы
// и получает готовый экземпляр Payment в ответ.
func (k *Kassa) SendPaymentConfig(config *yookassa.PaymentConfig) (*yookassa.Payment, error){
	paymentBytes, err := json.Marshal(config)
	if err != nil{
		return nil, err
	}

	resp, err := k.sendPostRequest(yookassa.PaymentsEndpoint, paymentBytes)
	if err != nil{
		return nil, err
	}

	p, err := k.handleResponse(resp)
	if err != nil{
		return nil, err
	}

	return p, nil
}

// GetPaymentInfo получает объект Payment по ID.
func (k *Kassa) GetPaymentInfo(id string)(*yookassa.Payment, error){
	resp, err := k.sendPostRequest(yookassa.PaymentsEndpoint + id, nil)
	if err != nil{
		return nil, err
	}

	p, err := k.handleResponse(resp)
	if err != nil{
		return nil, err
	}

	return p, nil
}

// handleResponse парсит ответ в экземпляр Payment.
func (k *Kassa) handleResponse(resp *http.Response)(*yookassa.Payment, error){
	var responseBytes []byte
	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil{
		return nil, err
	}

	p := yookassa.Payment{}
	err = json.Unmarshal(responseBytes, &p)
	if err != nil{
		return nil, err
	}

	return &p, nil
}

// sendPostRequest отправляет стандартный POST запрос с требуемыми настройками.
func (k *Kassa) sendPostRequest(endpoint string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, k.endpoint+endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(k.ShopID, k.Token)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotence-Key", UUIDGen())

	resp, err := k.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// sendPostRequest отправляет стандартный GET запрос с требуемыми настройками.
func (k *Kassa) sendGetRequest(endpoint string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, k.endpoint+endpoint, body)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(k.ShopID, k.Token)

	resp, err := k.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}