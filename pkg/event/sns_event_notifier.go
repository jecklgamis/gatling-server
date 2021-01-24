package event

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/jecklgamis/gatling-server/pkg/jsonutil"
	"log"
)

type SNSEventNotifier struct {
	ConfigMap map[string]string
	sns       snsiface.SNSAPI
}

func NewSNSEventNotifier(sns snsiface.SNSAPI, configMap map[string]string) *SNSEventNotifier {
	if _, ok := configMap["topicArn"]; !ok {
		log.Println("no topicArn found in config map")
		return nil
	}
	return &SNSEventNotifier{ConfigMap: configMap, sns: sns}
}

func (h *SNSEventNotifier) Event(event interface{}) {
	h.notify(event)
}

func CreateSNS(region string) *sns.SNS {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		log.Println("Failed creating AWS session", err)
		return nil
	}
	return sns.New(sess)
}

func (h *SNSEventNotifier) notify(event interface{}) error {
	client := h.sns
	input := &sns.PublishInput{
		Message:  aws.String(jsonutil.ToJson(event)),
		TopicArn: aws.String(h.ConfigMap["topicArn"]),
	}
	_, err := client.Publish(input)
	if err != nil {
		log.Println("Failed sending SNS message : ", err)
		return err
	}
	log.Println("Sent SNS message", jsonutil.ToJson(event))
	return nil
}
