// Copyright © 2012 Popog
package glml

var ContextSettingsDefault = ContextSettings{
	MajorVersion: 2,
	MinorVersion: 0,
}

type ContextSettings struct {
	DepthBits, // Bits of the depth buffer
	StencilBits, // Bits of the stencil buffer
	AntialiasingLevel, // Level of antialiasing
	MajorVersion, // Major number of the context version to create
	MinorVersion uint // Minor number of the context version to create
}

func absdif(a, b uint) uint {
	if a < b {
		return b - a
	}
	return a - b
}

func EvaluateFormat(rBitsPerPixel uint, rSettings ContextSettings, colorBits, depthBits, stencilBits, antialiasing uint) uint {
	return absdif(rBitsPerPixel, colorBits) +
		absdif(rSettings.DepthBits, depthBits) +
		absdif(rSettings.StencilBits, stencilBits) +
		absdif(rSettings.AntialiasingLevel, antialiasing)
}
