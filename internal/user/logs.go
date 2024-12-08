package user

import "log/slog"

func logAttrOrgID(orgID string) slog.Attr {
	return slog.String("orgID", orgID)
}

func logAttrUserID(userID string) slog.Attr {
	return slog.String("userID", userID)
}

func logAttrUser(user any) slog.Attr {
	return slog.Any("user", user)
}

func logAttrPathID(pathID string) slog.Attr {
	return slog.String("pathID", pathID)
}

func logAttrUsersLen(len int) slog.Attr {
	return slog.Int("usersLen", len)
}
