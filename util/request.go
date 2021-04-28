package util

import (
	"github.com/bwmarrin/discordgo"
)

/*
 * mute	boolean	whether the user is muted in voice channels. Will throw a 400 if the user is not in a voice channel	MUTE_MEMBERS
 * deaf	boolean	whether the user is deafened in voice channels. Will throw a 400 if the user is not in a voice channel	DEAFEN_MEMBERS
 * channel_id	snowflake	id of channel to move user to (if they are connected to voice)	MOVE_MEMBERS
 */

type ModifyVoiceStateParam struct {
	Mute		bool	`json:"mute"`
	Deaf		bool	`json:"deaf"`
	ChannelID	string	`json:"channel_id,omitempty"`
}

func RequestModifyVoiceState(s *discordgo.Session, guildID, userID string, mute, deaf bool, destination string) error {
	p := ModifyVoiceStateParam{
		Mute:	mute,
		Deaf:	deaf,
		ChannelID:	destination,
	}
	_, err := s.RequestWithBucketID("PATCH", discordgo.EndpointGuildMember(guildID, userID), p, discordgo.EndpointGuildMember(guildID, ""))
	return err
}


