//nolint:lll //hardcoded wow phrases
package wow

import (
	"context"
	"math/rand/v2"
)

type Query interface {
	GetNextWoW(ctx context.Context) (string, error)
}

type query struct {
	phrases []string
}

func (q *query) GetNextWoW(_ context.Context) (string, error) {
	// Using hardcoded list of phrases
	// We may want to change it to go to some API to get a next phrase
	// and fallback to the below if it failed. Or store them in a DB...

	nextIndex := rand.IntN(len(q.phrases)) //nolint:gosec //math/rand is ok here
	return q.phrases[nextIndex], nil
}

func NewQuery() Query {
	return &query{
		phrases: []string{
			"You create your own opportunities. Success doesn’t just come and find you–you have to go out and get it.",
			"Never break your promises. Keep every promise; it makes you credible.",
			"You are never as stuck as you think you are. Success is not final, and failure isn’t fatal.",
			"Happiness is a choice. For every minute you are angry, you lose 60 seconds of your own happiness.",
			"Habits develop into character. Character is the result of our mental attitude and the way we spend our time.",
			"Be happy with who you are. Being happy doesn’t mean everything is perfect but that you have decided to look beyond the imperfections.",
			"Don’t seek happiness–create it. You don’t need life to go your way to be happy.",
			"If you want to be happy, stop complaining. If you want happiness, stop complaining about how your life isn’t what you want and make it into what you do want.",
			"Asking for help is a sign of strength. Don’t let your fear of being judged stop you from asking for help when you need it. Sometimes asking for help is the bravest move you can make. You don’t have to go it alone.",
			"Replace every negative thought with a positive one. A positive mind is stronger than a negative thought.",
			"Accept what is, let go of what was, have faith in what will be. Sometimes you have to let go to let new things come in.",
			"A mind that is stretched by a new experience can never go back to what it was. Experience is what causes a person to make new mistakes instead of old ones.",
			"If you are not willing to learn, no one can help you. If you are determined to learn no one can stop you.",
			"Be confident enough to encourage confidence in others. Show those around you that you have confidence in them.",
			"Allow others to figure things out for themselves. The fixer ends up becoming the enabler. Let people figure it out for themselves; give them a chance to learn.",
			"Confidence is essential for a successful life. Don’t compare yourself to others; compare yourself to the person you were yesterday and give yourself the confidence to be better tomorrow.",
			"Admit your mistakes and don’t repeat them. If you can’t admit your mistakes, you are destined to repeat them.",
			"Be kind to yourself and forgive yourself. You can’t know what you haven’t yet learned, you can’t become proficient without first being a beginner and you can’t be perfect. Welcome to the human race.",
			"Failures are lessons in progress. Failure is always forgivable if you have the courage to learn its lessons and move forward in a new way.",
			"Make amends with those who have wronged you. Apologizing doesn’t always mean that you’re wrong and the other person is right. It just means that you value your relationships more than your ego.",
		},
	}
}
