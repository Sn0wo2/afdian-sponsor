package main

// Sponsor holds all the information about a sponsor.
type Sponsor struct {
	// Data fields
	Name         string
	OriginalName string
	Avatar       string

	// Sort fields
	AllSumAmount float64 // safe(?)
	LastPayTime  int

	// SVG rendering fields
	Index          int
	CenterX        int
	CenterY        int
	TextY          int
	Radius         int
	AvatarSize     int
	ImgMime        string
	ImgB64         string
	AnimationDelay float32
	Opacity        float32
	IsActive       bool
	TranslateX     int
	TranslateY     int
}
