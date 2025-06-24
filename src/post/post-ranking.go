package post

import (
	"agora/src/log"
	"math"
	"time"
)

// TODO: synchronous ranking, maybe parallelize this later
func (ph *PostHandler) GenerateNewRanks() {
	posts, err := ph.QueryAllPostsForRanking()
	if err != nil {
		log.Error.Printf("Could not retrieve posts: %v", err)
		return
	}

	for _, post := range posts {
		score := ph.calculatePostRank(post.FNrOfVotes, post.CreatedAt)
		log.Debug.Printf("Post ID: %d, Votes: %d, CreatedAt: %s, Rank: %d", post.ID, post.FNrOfVotes, post.CreatedAt, score)
		err = ph.UpdateRank(post.ID, score)
		if err != nil {
			log.Error.Printf("Could not update rank for post %d: %v", post.ID, err)
		}
	}
}

// score = (votes - 1) / (age in hours + damper)^gravity
// gravity = 1.8
// damper = 2

const gravity = 1
const damper = 2
const selfVotePenalty = 0
const factor = 100

func (ph *PostHandler) calculatePostRank(votes int, creationDate string) int {
	// Calculate the age of the post in hours
	creationTime, err := time.Parse(time.RFC3339, creationDate)
	if err != nil {
		log.Error.Printf("Could not parse creation date: %v", err)
		return 0
	}
	ageInHours := int(time.Since(creationTime).Hours())

	// Calculate the score
	numerator := float64(votes - selfVotePenalty)
	denominator := math.Pow(float64(ageInHours+damper), gravity)
	score := factor * (numerator / denominator)

	log.Debug.Printf("score='%v' numerator='%v' denominator='%v' ageinhours='%v' \n", score, numerator, denominator, ageInHours)
	return int(score)
}
