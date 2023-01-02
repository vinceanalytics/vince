package vince

import "strings"

func parseBotUA(ua string) *botMatch {
	fast := strings.ToLower(ua)
	// fast path with exact matches
	if m, ok := botsExactMatchMap[fast]; ok {
		return &botMatch{
			name:         m.name,
			category:     m.category,
			url:          m.url,
			producerName: m.producerName,
			producerURL:  m.producerURL,
		}
	}
	if allBotsReStandardMatch.MatchString(ua) {
		for _, m := range botsReList {
			if m.re.MatchString(ua) {
				return &botMatch{
					name:         m.name,
					category:     m.category,
					url:          m.url,
					producerName: m.producerName,
					producerURL:  m.producerURL,
				}
			}
		}
		return nil
	}
	if ok, _ := allBotsRe2Match.MatchString(ua); ok {
		for _, m := range botsRe2List {
			if ok, _ := m.re.MatchString(ua); ok {
				return &botMatch{
					name:         m.name,
					category:     m.category,
					url:          m.url,
					producerName: m.producerName,
					producerURL:  m.producerURL,
				}
			}
		}
	}
	return nil
}
