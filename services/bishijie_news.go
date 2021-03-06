package services

import (
	"context"
	"encoding/base64"
	"log"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/crossle/bishijie-news-mixin-bot/config"
	"github.com/crossle/bishijie-news-mixin-bot/models"
	"github.com/jasonlvhit/gocron"
)

type NewsService struct{}

type Stats struct {
	prevStoryId int
}

func (self Stats) getPrevTopStoryId() int {
	return self.prevStoryId
}

func (self *Stats) updatePrevTopStoryId(id int) {
	self.prevStoryId = id
}

func getTopStory() NewsFlash {
	stories, _ := GetStories()
	return stories[len(stories)-1]
}

func sendTopStoryToChannel(ctx context.Context, stats *Stats) {
	prevStoryId := stats.getPrevTopStoryId()
	stories, _ := GetStories()
	for i := len(stories) - 1; i >= 0; i-- {
		story := stories[i]
		println(story.Content)
		if story.ID > prevStoryId {
			log.Printf("Sending top story to channel...")
			stats.updatePrevTopStoryId(story.ID)
			subscribers, _ := models.FindSubscribers(ctx)
			for _, subscriber := range subscribers {
				conversationId := bot.UniqueConversationId(config.MixinClientId, subscriber.UserId)
				data := base64.StdEncoding.EncodeToString([]byte(story.Content))
				bot.PostMessage(ctx, conversationId, subscriber.UserId, bot.UuidNewV4().String(), "PLAIN_TEXT", data, config.MixinClientId, config.MixinSessionId, config.MixinPrivateKey)
			}
		} else {
			log.Printf("Same top story ID: %d, no message sent.", prevStoryId)
		}
	}
}
func (service *NewsService) Run(ctx context.Context) error {
	topStory := getTopStory()
	stats := &Stats{topStory.ID}
	gocron.Every(5).Minutes().Do(sendTopStoryToChannel, ctx, stats)
	<-gocron.Start()
	return nil
}
