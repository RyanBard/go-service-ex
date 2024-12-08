package org

import "log/slog"

func logAttrOrgID(orgID string) slog.Attr {
	return slog.String("orgID", orgID)
}

func logAttrOrgName(orgName string) slog.Attr {
	return slog.String("orgName", orgName)
}

func logAttrOrg(org any) slog.Attr {
	return slog.Any("org", org)
}

func logAttrPathID(pathID string) slog.Attr {
	return slog.String("pathID", pathID)
}

func logAttrOrgsLen(len int) slog.Attr {
	return slog.Int("orgsLen", len)
}
