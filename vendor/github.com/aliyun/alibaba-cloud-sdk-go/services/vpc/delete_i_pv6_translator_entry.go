package vpc

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//
// Code generated by Alibaba Cloud SDK Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
)

// DeleteIPv6TranslatorEntry invokes the vpc.DeleteIPv6TranslatorEntry API synchronously
// api document: https://help.aliyun.com/api/vpc/deleteipv6translatorentry.html
func (client *Client) DeleteIPv6TranslatorEntry(request *DeleteIPv6TranslatorEntryRequest) (response *DeleteIPv6TranslatorEntryResponse, err error) {
	response = CreateDeleteIPv6TranslatorEntryResponse()
	err = client.DoAction(request, response)
	return
}

// DeleteIPv6TranslatorEntryWithChan invokes the vpc.DeleteIPv6TranslatorEntry API asynchronously
// api document: https://help.aliyun.com/api/vpc/deleteipv6translatorentry.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DeleteIPv6TranslatorEntryWithChan(request *DeleteIPv6TranslatorEntryRequest) (<-chan *DeleteIPv6TranslatorEntryResponse, <-chan error) {
	responseChan := make(chan *DeleteIPv6TranslatorEntryResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DeleteIPv6TranslatorEntry(request)
		if err != nil {
			errChan <- err
		} else {
			responseChan <- response
		}
	})
	if err != nil {
		errChan <- err
		close(responseChan)
		close(errChan)
	}
	return responseChan, errChan
}

// DeleteIPv6TranslatorEntryWithCallback invokes the vpc.DeleteIPv6TranslatorEntry API asynchronously
// api document: https://help.aliyun.com/api/vpc/deleteipv6translatorentry.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DeleteIPv6TranslatorEntryWithCallback(request *DeleteIPv6TranslatorEntryRequest, callback func(response *DeleteIPv6TranslatorEntryResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DeleteIPv6TranslatorEntryResponse
		var err error
		defer close(result)
		response, err = client.DeleteIPv6TranslatorEntry(request)
		callback(response, err)
		result <- 1
	})
	if err != nil {
		defer close(result)
		callback(nil, err)
		result <- 0
	}
	return result
}

// DeleteIPv6TranslatorEntryRequest is the request struct for api DeleteIPv6TranslatorEntry
type DeleteIPv6TranslatorEntryRequest struct {
	*requests.RpcRequest
	ResourceOwnerId       requests.Integer `position:"Query" name:"ResourceOwnerId"`
	Ipv6TranslatorEntryId string           `position:"Query" name:"Ipv6TranslatorEntryId"`
	ResourceOwnerAccount  string           `position:"Query" name:"ResourceOwnerAccount"`
	ClientToken           string           `position:"Query" name:"ClientToken"`
	OwnerAccount          string           `position:"Query" name:"OwnerAccount"`
	Ipv6TranslatorId      string           `position:"Query" name:"Ipv6TranslatorId"`
	OwnerId               requests.Integer `position:"Query" name:"OwnerId"`
}

// DeleteIPv6TranslatorEntryResponse is the response struct for api DeleteIPv6TranslatorEntry
type DeleteIPv6TranslatorEntryResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
}

// CreateDeleteIPv6TranslatorEntryRequest creates a request to invoke DeleteIPv6TranslatorEntry API
func CreateDeleteIPv6TranslatorEntryRequest() (request *DeleteIPv6TranslatorEntryRequest) {
	request = &DeleteIPv6TranslatorEntryRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Vpc", "2016-04-28", "DeleteIPv6TranslatorEntry", "vpc", "openAPI")
	return
}

// CreateDeleteIPv6TranslatorEntryResponse creates a response to parse from DeleteIPv6TranslatorEntry response
func CreateDeleteIPv6TranslatorEntryResponse() (response *DeleteIPv6TranslatorEntryResponse) {
	response = &DeleteIPv6TranslatorEntryResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
