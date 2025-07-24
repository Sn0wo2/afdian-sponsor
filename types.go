package main

type sponsor struct {
	// Data fields
	Name   string
	Avatar string

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
