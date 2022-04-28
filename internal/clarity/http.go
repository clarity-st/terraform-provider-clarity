package clarity

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	Authorization = "Authorization"
)

//lowtech
func (config *Client) do(method string, path string, payload io.Reader) (int, []byte, error) {
	endpoint := fmt.Sprintf("%s/%s", config.Host, path)
	req, err := http.NewRequest(method, endpoint, payload)
	if err != nil {
		return 0, nil, err
	}

	req.Header.Set(Authorization, fmt.Sprintf("Bearer %s", config.Token))

	rsp, err := config.Client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer rsp.Body.Close()

	output, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return 0, nil, err
	}

	return rsp.StatusCode, output, nil

}
