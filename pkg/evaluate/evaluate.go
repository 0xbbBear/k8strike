package evaluate

import "log"

func CallBasics() {
	if err := NewEvaluator().RunProfile(ProfileBasic, nil); err != nil {
		log.Printf("basic evaluation failed: %v", err)
	}
}

func CallAddedFunc() {
	if err := NewEvaluator().RunProfile(ProfileAdditional, nil); err != nil {
		log.Printf("additional evaluation failed: %v", err)
	}
}
