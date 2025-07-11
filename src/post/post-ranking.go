package post

import (
	"agora/src/log"
	"fmt"
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
		err = ph.UpdateRank(post.ID, score)
		if err != nil {
			log.Error.Printf("Could not update rank for post %d: %v", post.ID, err)
		}
	}
}

func (ph *PostHandler) GenerateNewRanksForPost(postID int) error {
	post, err := ph.QueryOnePost(postID)
	if err != nil {
		return fmt.Errorf("could not retrieve post %d: %v", postID, err)
	}

	newRank := ph.calculatePostRank(post.FNrOfVotes, post.CreatedAt)
	ph.UpdateRank(post.ID, newRank)

	return nil
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

	return int(score)
}
