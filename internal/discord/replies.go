package discord

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/jace-ys/hot-potato-discord/internal/hotpotato"
)

const (
	MessageFlagEphemeral = 1 << 6
)

type Reply struct {
	Message   string
	Embed     *discordgo.MessageEmbed
	GIF       *GIF
	Ephemeral bool
}

type GIF struct {
	Name string
	URL  string
}

var ExplodeGIFs = []*GIF{
	{Name: "explode-1.gif", URL: "https://media.giphy.com/media/3o7bu4EJkrXG9Bvs9G/giphy.gif"},
	{Name: "explode-2.gif", URL: "https://media.giphy.com/media/KbeMMTRmTCF11CPLq1/giphy.gif"},
	{Name: "explode-3.gif", URL: "https://media.giphy.com/media/AikDcN9ZxLoDhE1v9A/giphy.gif"},
	{Name: "explode-4.gif", URL: "https://media.giphy.com/media/jPMXMZrhNxOK59b5rp/giphy.gif"},
	{Name: "explode-5.gif", URL: "https://media.giphy.com/media/VFMd7k7TRABLtuLRaY/giphy.gif"},
	{Name: "explode-6.gif", URL: "https://media.giphy.com/media/eFifJQ2SUYxO0/giphy.gif"},
}

func RandomExplodeGIF() *GIF {
	i := rand.Intn(len(ExplodeGIFs))
	return ExplodeGIFs[i]
}

func TossSuccessReply(actorUserID, targetUserID string, rsp *hotpotato.TossResponse) *Reply {
	reply := &Reply{}

	var sb strings.Builder
	if rsp.Turn == 1 {
		if actorUserID == targetUserID {
			sb.WriteString(fmt.Sprintf("<@!%s> grabbed a **%s** fresh out of the oven but forgot to toss it! 🙈", actorUserID, rsp.Potato))
		} else {
			sb.WriteString(fmt.Sprintf("<@!%s> grabbed a **%s** fresh out of the oven and tossed it to <@%s>!", actorUserID, rsp.Potato, targetUserID))
		}
	} else {
		if actorUserID == targetUserID {
			sb.WriteString(fmt.Sprintf("<@!%s> tried to juggle the **%s** like a fool 😵‍💫", targetUserID, rsp.Potato))
		} else {
			sb.WriteString(fmt.Sprintf("<@!%s> tossed the **%s** to <@!%s>!", actorUserID, rsp.Potato, targetUserID))
		}
	}

	if rsp.Exploded {
		sb.WriteString(fmt.Sprintf("\nOh no, the **%s** exploded in <@!%s>'s face! 🤢", rsp.Potato, targetUserID))
		reply.GIF = RandomExplodeGIF()
	}

	reply.Message = sb.String()
	return reply
}

func TossInvalidTargetReply(targetUserID string) *Reply {
	return &Reply{
		Message:   fmt.Sprintf("You can't toss a potato to <@!%s>. Try someone else!", targetUserID),
		Ephemeral: true,
	}
}

func TossNotHolderReply(holderUserID string) *Reply {
	return &Reply{
		Message:   fmt.Sprintf("You can't toss the potato as <@!%s> is currently holding it!", holderUserID),
		Ephemeral: true,
	}
}

func StealSuccessReply(actorUserID, targetUserID string, rsp *hotpotato.StealResponse) *Reply {
	reply := &Reply{}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<@!%s> stole the **%s** from <@!%s>!", actorUserID, rsp.Potato, targetUserID))

	if rsp.Exploded {
		sb.WriteString(fmt.Sprintf("\nOh no, the **%s** exploded in <@!%s>'s face! 🤢", rsp.Potato, actorUserID))
		reply.GIF = RandomExplodeGIF()
	}

	reply.Message = sb.String()
	return reply
}

func StealInvalidTargetReply(targetUserID string) *Reply {
	return &Reply{
		Message:   fmt.Sprintf("You can't steal a potato from <@!%s>. Try someone else!", targetUserID),
		Ephemeral: true,
	}
}

func StealNotHolderReply(targetUserID, holderUserID string) *Reply {
	return &Reply{
		Message:   fmt.Sprintf("You can't steal the potato from <@!%s> as <@!%s> is currently holding it!", targetUserID, holderUserID),
		Ephemeral: true,
	}
}

func CookSuccessReply(actorUserID string, rsp *hotpotato.CookResponse) *Reply {
	reply := &Reply{}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<@!%s> cooked the **%s** and made it hotter! 🔥", actorUserID, rsp.Potato))

	if rsp.Exploded {
		sb.WriteString(fmt.Sprintf("\nOh no, the **%s** exploded in <@!%s>'s face! 🤢", rsp.Potato, actorUserID))
		reply.GIF = RandomExplodeGIF()
	}

	reply.Message = sb.String()
	return reply
}

func CookNotHolderReply(holderUserID string) *Reply {
	return &Reply{
		Message:   fmt.Sprintf("You can't cook the potato as <@!%s> is currently holding it!", holderUserID),
		Ephemeral: true,
	}
}

func WhereSuccessReply(rsp *hotpotato.GetHolderResponse) *Reply {
	return &Reply{
		Message: fmt.Sprintf("The **%s** is currently being held by <@!%s>", rsp.Potato, rsp.HolderUserID),
	}
}

func LeaderboardSuccessReply(rsp *hotpotato.GetLeaderboardResponse) *Reply {
	var sb strings.Builder
	sb.WriteString("**🥁 __Deaths by Hot 🔥 Potato 🥔 Leaderboard__ 🥁**")
	sb.WriteString("\nHere are the top 10 losers who have had the most hot potatoes explode in their faces 🤢")
	sb.WriteString("\n")

	if len(rsp.Leaderboard) == 0 {
		sb.WriteString("\n*😇 It seems like no one has died yet, time to start tossing some potatoes! 🔥🥔*")
		return &Reply{
			Message: sb.String(),
		}
	}

	for i, entry := range rsp.Leaderboard {
		ranking := i + 1

		var prefix string
		switch ranking {
		case 1:
			prefix = "🥇"
		case 2:
			prefix = "🥈"
		case 3:
			prefix = "🥉"
		case 4:
			prefix = "4️⃣"
		case 5:
			prefix = "5️⃣"
		case 6:
			prefix = "6️⃣"
		case 7:
			prefix = "7️⃣"
		case 8:
			prefix = "8️⃣"
		case 9:
			prefix = "9️⃣"
		case 10:
			prefix = "🔟"
		default:
			prefix = "❗️"
		}

		sb.WriteString(fmt.Sprintf("\n%s <@!%s> - %d deaths", prefix, entry.UserID, entry.Count))
	}

	return &Reply{
		Message: sb.String(),
	}
}

func NoOngoingGameReply() *Reply {
	return &Reply{
		Message:   "There doesn't seem to be an ongoing game in this channel. Start one by tossing a potato!",
		Ephemeral: true,
	}
}

func UnexpectedErrorReply() *Reply {
	return &Reply{
		Message:   "I am having difficulty processing your request right now. Please try again later.",
		Ephemeral: true,
	}
}

func (b *Bot) reply(s *discordgo.Session, i *discordgo.InteractionCreate, reply *Reply) error {
	ir := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: reply.Message,
		},
	}

	if reply.GIF != nil {
		ir.Data.Embeds = append(ir.Data.Embeds, &discordgo.MessageEmbed{
			Image: &discordgo.MessageEmbedImage{URL: reply.GIF.URL},
		})
	}

	if reply.Ephemeral {
		ir.Data.Flags = MessageFlagEphemeral
	}

	if err := s.InteractionRespond(i.Interaction, ir); err != nil {
		return fmt.Errorf("error responding to interaction: %w", err)
	}

	return nil
}
