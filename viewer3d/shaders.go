package viewer3d

import (
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/gogpu/naga"
	"github.com/gogpu/naga/ir"
	"github.com/gogpu/naga/msl"
	"github.com/gogpu/naga/spirv"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

// The shaders are compiled from WGSL when the viewer starts.
// SDL reads the vertex uniforms from descriptor set 1 in SPIR-V,
// so the uniforms are bound to group 1.

const mapShaderWGSL = `
struct Uniforms {
    mvp: mat4x4<f32>,
}
@group(1) @binding(0) var<uniform> uniforms: Uniforms;

@vertex
fn vs_main(@location(0) pos: vec3<f32>) -> @builtin(position) vec4<f32> {
    return uniforms.mvp * vec4<f32>(pos, 1.0);
}

@fragment
fn fs_main() -> @location(0) vec4<f32> {
    return vec4<f32>(0.5, 0.5, 0.5, 0.06);
}
`

////////////////////////////////////////////////////////////////////////////////

const entityShaderWGSL = `
struct Uniforms {
    mvp: mat4x4<f32>,
}
@group(1) @binding(0) var<uniform> uniforms: Uniforms;

struct VsOut {
    @builtin(position) pos: vec4<f32>,
    @location(0) color: vec4<f32>,
}

@vertex
fn vs_main(@location(0) pos: vec3<f32>, @location(1) color: vec4<f32>) -> VsOut {
    var out: VsOut;
    out.pos = uniforms.mvp * vec4<f32>(pos, 1.0);
    out.color = color;
    return out;
}

@fragment
fn fs_main(@location(0) color: vec4<f32>) -> @location(0) vec4<f32> {
    return color;
}
`

////////////////////////////////////////////////////////////////////////////////

// compileShaders compiles the `vs_main` and `fs_main` entry
// points of a WGSL shader into MSL for Metal or SPIR-V for
// Vulkan, whichever the device takes, and creates them. The
// vertex shader has one uniform buffer.
func compileShaders(device *sdl.GPUDevice, source string) (*sdl.GPUShader, *sdl.GPUShader, error) {

	//----------------------------------------------------------------------------//

	ast, err := naga.Parse(source)
	if err != nil {
		return nil, nil, errors.New(
			"failed to parse shader",
			errors.Error("error", err),
		)
	}

	module, err := naga.LowerWithSource(ast, source)
	if err != nil {
		return nil, nil, errors.New(
			"failed to lower shader",
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	var code []byte
	var format sdl.GPUShaderFormat
	vertexEntry, fragmentEntry := "vs_main", "fs_main"

	formats := device.ShaderFormats()

	switch {
	case formats&sdl.GPU_SHADERFORMAT_MSL != 0:

		// SDL binds the uniform buffers first in the Metal buffer
		// table, in the order of their slots
		slot := uint8(0)
		options := msl.DefaultOptions()
		options.PerEntryPointMap = map[string]msl.EntryPointResources{
			"vs_main": {Resources: map[ir.ResourceBinding]msl.BindTarget{
				{Group: 1, Binding: 0}: {Buffer: &slot},
			}},
			"fs_main": {Resources: map[ir.ResourceBinding]msl.BindTarget{}},
		}

		mslSource, info, err := msl.Compile(module, options)
		if err != nil {
			return nil, nil, errors.New(
				"failed to compile shader",
				errors.String("format", "msl"),
				errors.Error("error", err),
			)
		}

		code = []byte(mslSource)
		format = sdl.GPU_SHADERFORMAT_MSL
		vertexEntry = info.EntryPointNames["vs_main"]
		fragmentEntry = info.EntryPointNames["fs_main"]

	case formats&sdl.GPU_SHADERFORMAT_SPIRV != 0:

		code, err = naga.GenerateSPIRV(module, spirv.Options{Version: spirv.Version1_0})
		if err != nil {
			return nil, nil, errors.New(
				"failed to compile shader",
				errors.String("format", "spirv"),
				errors.Error("error", err),
			)
		}

		format = sdl.GPU_SHADERFORMAT_SPIRV

	default:
		return nil, nil, errors.New(
			"gpu device takes no supported shader format",
			errors.String("driver", device.Driver()),
		)
	}

	//----------------------------------------------------------------------------//

	vertex, err := device.CreateGPUShader(&sdl.GPUShaderCreateInfo{
		Code:              code,
		Entrypoint:        vertexEntry,
		Format:            format,
		Stage:             sdl.GPU_SHADERSTAGE_VERTEX,
		NumUniformBuffers: 1,
	})
	if err != nil {
		return nil, nil, errors.New(
			"failed to create vertex shader",
			errors.Error("error", err),
		)
	}

	fragment, err := device.CreateGPUShader(&sdl.GPUShaderCreateInfo{
		Code:       code,
		Entrypoint: fragmentEntry,
		Format:     format,
		Stage:      sdl.GPU_SHADERSTAGE_FRAGMENT,
	})
	if err != nil {
		device.ReleaseShader(vertex)
		return nil, nil, errors.New(
			"failed to create fragment shader",
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	return vertex, fragment, nil

	//----------------------------------------------------------------------------//
}
