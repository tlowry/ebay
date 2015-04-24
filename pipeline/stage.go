package pipeline

import (
	"appengine"
	"time"
)

type EbayItem struct {
	Description string
	Url         string
	ImageUrl    string
	ListingId   string
	Currency    string
	Price       float64
	Tier        string
	ExpiryDate  time.Time
}

func (st EbayItem) String() string {
	return "EbayItem, desc=" + st.Description
}

type StageIF interface {
	Init()
	Run()
	SetIn(In chan EbayItem)
	GetIn() chan EbayItem
	SetOut(In chan EbayItem)
	GetOut() chan EbayItem
}

type Stage struct {
	StageIF
	name string
	In   chan EbayItem
	Out  chan EbayItem
	ctx  appengine.Context
}

func (st *Stage) GetContext() appengine.Context {
	return st.ctx
}

func NewStage(ctx appengine.Context) *Stage {
	st := Stage{}
	st.ctx = ctx
	return &st
}

func (st *Stage) SetIn(in chan EbayItem) {
	st.In = in
}

func (st *Stage) SetOut(out chan EbayItem) {
	st.Out = out
}

func (st *Stage) GetIn() chan EbayItem {
	return st.In
}

func (st *Stage) GetOut() chan EbayItem {
	return st.Out
}

func (st *Stage) SetName(name string) {
	st.name = name
}

func (st *Stage) GetName() string {
	st.ctx.Infof("Stage::GetName" + st.name)
	return st.name
}

func (st *Stage) Init() {
}

type StageGroup struct {
	StageIF
	In        *chan EbayItem
	Out       *chan EbayItem
	instances []StageIF
}

func NewStageGroup() *StageGroup {
	group := StageGroup{}
	return &group
}

func (st *StageGroup) SetIn(in chan EbayItem) {
	st.In = &in
}

func (st *StageGroup) SetOut(out chan EbayItem) {
	st.Out = &out
}

func (st *StageGroup) GetIn() chan EbayItem {
	return *st.In
}

func (st *StageGroup) GetOut() chan EbayItem {
	return *st.Out
}

func (st *StageGroup) AddInstance(instance StageIF) {
	st.instances = append(st.instances, instance)
}

func (st *StageGroup) Init() {
	for _, s := range st.instances {
		s.SetOut(*st.Out)
		s.Init()
	}
}

func (st *StageGroup) Run() {
	for _, st := range st.instances {
		go st.Run()
	}
}
