package pipeline

import (
	"appengine"
	"sync"
)

type Pipeline struct {
	stages []StageIF
	lock   sync.Mutex
	ctx    appengine.Context
}

func NewPipeline(ctx appengine.Context) *Pipeline {
	pipe := Pipeline{}
	pipe.ctx = ctx
	pipe.stages = make([]StageIF, 0)

	return &pipe
}
func (pipe *Pipeline) GetContext() appengine.Context {
	return pipe.ctx
}
func (pipe *Pipeline) AddBack(stage StageIF, bufSize int) {
	pipe.lock.Lock()
	stageNum := len(pipe.stages)

	if stageNum > 0 {
		pipe.ctx.Infof("Connecting previous stage to new")
		lastStage := pipe.stages[stageNum-1]
		linkChan := lastStage.GetOut()
		stage.SetIn(linkChan)
	} else {
		pipe.ctx.Infof("Creating first stage")
	}

	outChan := make(chan EbayItem, bufSize)
	stage.SetOut(outChan)

	pipe.stages = append(pipe.stages, stage)
	pipe.lock.Unlock()
}

func (pipe *Pipeline) Init() {
	for _, stage := range pipe.stages {
		stage.Init()
	}
}

func (pipe *Pipeline) Run() {
	for _, stage := range pipe.stages {
		go stage.Run()
	}
}
