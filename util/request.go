package util

import (
	"github.com/bwmarrin/discordgo"
)

/*
 * mute	boolean	whether the user is muted in voice channels. Will throw a 400 if the user is not in a voice channel	MUTE_MEMBERS
 * deaf	boolean	whether the user is deafened in voice channels. Will throw a 400 if the user is not in a voice channel	DEAFEN_MEMBERS
 * channel_id	snowflake	id of channel to move user to (if they are connected to voice)	MOVE_MEMBERS
 */

type ModifyVSParams struct {
	Mute		*bool	`json:"mute,omitempty"`
	Deaf		*bool	`json:"deaf,omitempty"`
	ChannelID	*string	`json:"channel_id,omitempty"`
}

type ModifyVSParam func(*ModifyVSParams)

func ModifyVSParamMute(mute bool) ModifyVSParam {
	return func(m *ModifyVSParams) {
		m.Mute = &mute
	}
}

func ModifyVSParamDeaf(deaf bool) ModifyVSParam {
	return func(m *ModifyVSParams) {
		m.Deaf = &deaf
	}
}

func ModifyVSParamChannelID(channelID string) ModifyVSParam {
	return func(m *ModifyVSParams) {
		m.ChannelID = &channelID
	}
}

func RequestModifyVS(s *discordgo.Session, guildID, userID string, params ...ModifyVSParam) error {
	p := &ModifyVSParams{}

	for _, param := range params {
		param(p)
	}

	_, err := s.RequestWithBucketID("PATCH", discordgo.EndpointGuildMember(guildID, userID), *p, discordgo.EndpointGuildMember(guildID, ""))
	return err
}
