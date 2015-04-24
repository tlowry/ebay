package pipeline

import (
	"log"
	"sync"
)

type Pipeline struct {
	stages []StageIF
	lock   sync.Mutex
}

func NewPipeline() *Pipeline {
	pipe := Pipeline{}
	pipe.stages = make([]StageIF, 0)
	return &pipe
}

func (this *Pipeline) AddBack(stage StageIF) {
	this.lock.Lock()
	stageNum := len(this.stages)

	if stageNum > 0 {
		log.Println("Connecting previous stage to new")
		lastStage := this.stages[stageNum-1]
		linkChan := lastStage.GetOut()
		stage.SetIn(linkChan)
	} else {
		log.Println("Creating first stage")
	}

	outChan := make(chan EbayItem)
	stage.SetOut(outChan)

	this.stages = append(this.stages, stage)
	this.lock.Unlock()
}

func (this *Pipeline) Init() {
	for _, stage := range this.stages {
		stage.Init()
	}
}

func (this *Pipeline) Run() {
	for _, stage := range this.stages {
		go stage.Run()
	}
}
