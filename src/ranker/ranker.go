package ranker

import (
	"agora/src/log"
	"agora/src/post"
	"time"
)

type Ranker struct {
	hourTicker *time.Ticker
	ph         *post.PostHandler // Assuming PostHandler is defined in the same package or imported
}

func NewRanker(ph *post.PostHandler) *Ranker {
	return &Ranker{
		hourTicker: time.NewTicker(time.Second * 10),
		ph:         ph,
	}
}

func (r *Ranker) RankPosts() {
	r.ph.GenerateNewRanks()
}

func (r *Ranker) Start() {
	r.RankPosts()

	go func() {
		for {
			if r.hourTicker == nil {
				log.Error.Println("Ranker ticker channel is nil, stopping ranker service.")
				break
			}

			<-r.hourTicker.C
			log.Debug.Println("running post ranking")
			r.RankPosts()
		}
	}()

}

func (r *Ranker) Stop() {
	r.hourTicker.Stop()
}
