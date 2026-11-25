package twitch

const userFields = `id login displayName profileImageURL(width: 300)`

const gameFields = `id displayName slug boxArtURL(width: 285, height: 380) viewersCount broadcastersCount`

const streamFields = `id title viewersCount createdAt previewImageURL(width: 640, height: 360)
	game { ` + gameFields + ` }
	freeformTags { name }`

const streamWithBroadcaster = streamFields + `
	broadcaster { ` + userFields + ` broadcastSettings { title language } }`

const videoFields = `id title lengthSeconds viewCount createdAt broadcastType status animatedPreviewURL previewThumbnailURL(width: 640, height: 360)
	game { ` + gameFields + ` }`

const clipFields = `slug title durationSeconds viewCount createdAt videoOffsetSeconds thumbnailURL(width: 480, height: 272)
	game { ` + gameFields + ` }
	video { id }`

const queryTopStreams = `query TopStreams($languages: [Language!], $first: Int, $after: Cursor) {
	streams(first: $first, after: $after, options: { sort: VIEWER_COUNT, broadcasterLanguages: $languages }) {
		edges { cursor node { ` + streamWithBroadcaster + ` } }
		pageInfo { hasNextPage }
	}
}`

const queryTagStreams = `query TagStreams($tags: [String!], $languages: [Language!], $first: Int, $after: Cursor) {
	streams(first: $first, after: $after, options: { sort: VIEWER_COUNT, freeformTags: $tags, broadcasterLanguages: $languages }) {
		edges { cursor node { ` + streamWithBroadcaster + ` } }
		pageInfo { hasNextPage }
	}
}`

const queryTopGames = `query TopGames($first: Int, $after: Cursor) {
	games(first: $first, after: $after) {
		edges { cursor node { ` + gameFields + ` } }
		pageInfo { hasNextPage }
	}
}`

const queryGameStreams = `query GameStreams($slug: String, $languages: [Language!], $first: Int, $after: Cursor) {
	game(slug: $slug) {
		` + gameFields + `
		streams(first: $first, after: $after, options: { sort: VIEWER_COUNT, broadcasterLanguages: $languages }) {
			edges { cursor node { ` + streamWithBroadcaster + ` } }
			pageInfo { hasNextPage }
		}
	}
}`

const queryChannel = `query Channel($login: String) {
	user(login: $login, lookupType: ALL) {
		` + userFields + ` description bannerImageURL
		followers { totalCount }
		roles { isPartner }
		stream { ` + streamFields + ` archiveVideo { id createdAt } }
	}
}`

const queryChannelVideos = `query ChannelVideos($login: String, $first: Int, $after: Cursor) {
	user(login: $login, lookupType: ALL) {
		` + userFields + `
		videos(first: $first, after: $after, sort: TIME) {
			edges { cursor node { ` + videoFields + ` } }
			pageInfo { hasNextPage }
		}
	}
}`

const queryChannelClips = `query ChannelClips($login: String, $first: Int, $after: Cursor) {
	user(login: $login, lookupType: ALL) {
		` + userFields + `
		clips(first: $first, after: $after, criteria: { period: LAST_MONTH }) {
			edges { cursor node { ` + clipFields + ` } }
			pageInfo { hasNextPage }
		}
	}
}`

const queryVideo = `query Video($id: ID) {
	video(id: $id) {
		` + videoFields + `
		owner { ` + userFields + ` }
	}
}`

const queryClip = `query Clip($slug: ID!) {
	clip(slug: $slug) {
		` + clipFields + `
		broadcaster { ` + userFields + ` }
		assets { videoQualities { quality frameRate sourceURL } }
		playbackAccessToken(params: { platform: "web", playerType: "site" }) { signature value }
	}
}`

const querySearchChannels = `query SearchChannels($query: String!, $first: Int, $after: Cursor) {
	searchUsers(userQuery: $query, first: $first, after: $after) {
		edges { cursor node { ` + userFields + ` description followers { totalCount } stream { ` + streamFields + ` } } }
		pageInfo { hasNextPage }
	}
}`

const querySearchGames = `query SearchGames($query: String!, $first: Int, $after: Cursor) {
	searchCategories(query: $query, first: $first, after: $after) {
		edges { cursor node { ` + gameFields + ` } }
		pageInfo { hasNextPage }
	}
}`

const querySearchStreams = `query SearchStreams($query: String!, $first: Int, $after: Cursor) {
	searchStreams(userQuery: $query, first: $first, after: $after) {
		edges { cursor node { ` + streamWithBroadcaster + ` } }
		pageInfo { hasNextPage }
	}
}`

const queryFollowedLive = `query FollowedLive($first: Int, $after: Cursor) {
	user {
		followedLiveUsers(first: $first, after: $after) {
			edges { cursor node { ` + userFields + ` stream { ` + streamFields + ` } } }
			pageInfo { hasNextPage }
		}
	}
}`

const queryFollowedVideos = `query FollowedVideos($first: Int, $after: Cursor) {
	user {
		followedVideos(first: $first, after: $after, types: [ARCHIVE], sort: TIME) {
			edges { cursor node { ` + videoFields + ` owner { ` + userFields + ` } } }
			pageInfo { hasNextPage }
		}
	}
}`

const queryFollowedChannels = `query FollowedChannels($first: Int, $after: Cursor) {
	user {
		follows(first: $first, after: $after) {
			edges { cursor node { ` + userFields + ` } }
			pageInfo { hasNextPage }
		}
	}
}`

const queryCurrentUser = `query CurrentUser { currentUser { ` + userFields + ` } }`
