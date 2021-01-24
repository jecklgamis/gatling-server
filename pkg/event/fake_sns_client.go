package event

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
)

type fakeSNSClient struct {
	snsiface.SNSAPI
	input  *sns.PublishInput
	output *sns.PublishOutput
	error  error
}

func StringAddr(str string) *string {
	return &str
}

func NewNonFailingSNSClient() *fakeSNSClient {
	return &fakeSNSClient{
		output: &sns.PublishOutput{
			MessageId:      StringAddr("some-id"),
			SequenceNumber: StringAddr("some-seq-number")}}
}

func NewFailingSNSClient() *fakeSNSClient {
	return &fakeSNSClient{error: fmt.Errorf("some SNS error")}
}

func (f *fakeSNSClient) Publish(input *sns.PublishInput) (*sns.PublishOutput, error) {
	f.input = input

	if f.error != nil {
		return nil, f.error
	}
	return f.output, nil
}
