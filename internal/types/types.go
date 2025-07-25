package types

// Sponsor holds all the information about a sponsor.
type Sponsor struct {
	// Data fields
	Name         string
	OriginalName string
	Avatar       string

	// SVG rendering fields
	Index      int
	CenterX    int
	CenterY    int
	X          int
	Y          int
	TextY      int
	Radius     int
	AvatarSize int
	ImgMime    string
	ImgB64     string
}
