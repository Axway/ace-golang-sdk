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
	showMetadata("\t", aceMsg.GetMetaData())
	if aceMsg.BusinessMessage != nil && aceMsg.BusinessMessage.Payload != nil {
		log.Debugf("\tBusinessMessage.Payload.Body: '%v'", string(aceMsg.BusinessMessage.Payload.Body))
	} else if aceMsg.BusinessMessage != nil {
		log.Debugf("\tBusinessMessage.Payload: %v", aceMsg.BusinessMessage.Payload)
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
	log.Debugf("%s message has %d metadata item(s):", nestingTabs, len(md))
	for key, val := range md {
		log.Debugf("%s metadata key='%s' value='%s'", nestingTabs, key, val)
	}
}
