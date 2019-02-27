package util

import (
	"github.com/Axway/ace-golang-sdk/rpc"
	"github.com/Axway/ace-golang-sdk/util/logging"
)

var zlog = logging.Logger()

// Show - display entire ACE message using 'log' package
func Show(descr string, aceMsg *rpc.Message) {
	log := zlog.Sugar()

	log.Debugf("%s", descr)
	log.Debugf("\tUUID: '%s'", aceMsg.GetUUID())
	log.Debugf("\tParent_UUID: '%s'", aceMsg.GetParent_UUID())
	log.Debugf("\tCHN_UUID: '%s'", aceMsg.GetCHN_UUID())
	log.Debugf("\tCHX_UUID: '%s'", aceMsg.GetCHX_UUID())
	log.Debugf("\tSequenceTerm: %d", aceMsg.SequenceTerm)
	log.Debugf("\tSequenceUpperBound: %d", aceMsg.SequenceUpperBound)
	log.Debugf("\tPattern: %v, with %d children: ", aceMsg.Pattern, len(aceMsg.Pattern.Child))

	showPattern("\t\t", aceMsg.Pattern.Child)

	log.Debugf("\tTopicName: '%s'", aceMsg.GetTopicName())
	log.Debugf("\tID: '%s'", aceMsg.GetID())
	if aceMsg.ErrorType != rpc.Message_NONE {
		log.Debugf("\tError: %s", aceMsg.ErrorType)
		log.Debugf("\tError description: %s", aceMsg.GetErrorDescription())
	}
	log.Debugf("\tOpentracingContext:", aceMsg.GetOpentracingContext())
	if aceMsg.BusinessMessage != nil {
		showMetadata("\t business", aceMsg.BusinessMessage.GetMetaData())
		if aceMsg.BusinessMessage.Payload != nil {
			log.Debugf("\tBusinessMessage.Payload.Body (as string): '%s'", string(aceMsg.BusinessMessage.Payload.Body))
		}
	} else {
		log.Debugf("\tBusinessMessage: %v", aceMsg.BusinessMessage)
	}
	log.Debug()
}

func showPattern(nestingTabs string, child []*rpc.StepPattern) {
	if len(child) == 0 {
		return
	}
	for idx, stepPattern := range child {
		zlog.Sugar().Debugf("%sStepPattern[%d].ServiceName: '%s' len(Child):%d\n",
			nestingTabs, idx, stepPattern.ServiceName, len(stepPattern.Child))
		showPattern(nestingTabs+"\t", stepPattern.Child)
	}
}

func showMetadata(nestingTabs string, md map[string]string) {
	log := zlog.Sugar()
	if len(md) > 0 {
		log.Debugf("%s message has %d metadata item(s):", nestingTabs, len(md))
	} else {
		log.Debugf("%s message has no metadata items:", nestingTabs)
	}
	for key, val := range md {
		log.Debugf("%s metadata key='%s' value='%s'", nestingTabs, key, val)
	}
}
