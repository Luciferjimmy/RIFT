package types

// SystemProfile is the unified struct describing the host macOS machine.
// It is used for both local optimization and cloud AI edge functions.
type SystemProfile struct {
	Chip           string `json:"chip"`
	RAMGB          int    `json:"ramGB"`
	MacOSVersion   string `json:"macosVersion"`
	GPUMetalFamily string `json:"gpuMetalFamily"`
	HasGPTK        bool   `json:"hasGPTK"`
	HasDXVK        bool   `json:"hasDXVK"`
	HasMoltenVK    bool   `json:"hasMoltenVK"`
	HardwareUUID   string `json:"hardwareUUID"`

	// Additional fields from the optimizer pipeline
	AppleSilicon   bool `json:"apple_silicon"`
	MetalGPUFamily int  `json:"metal_gpu_family"` // Numeric representation
	Rosetta2       bool `json:"rosetta2"`
	HasWine        bool `json:"has_wine"`
}

// RuntimeConfig represents the final runtime engine + environment configuration
// used right before launch. It merges GameRuntimeConfig and hardware_hints.RuntimeConfig.
type RuntimeConfig struct {
	// Whisky Settings
	EnhancedSync string `json:"enhanced_sync"` // "none", "esync", "msync"
	MetalHUD bool `json:"metal_hud"`
	MetalTrace bool `json:"metal_trace"`
	DXREnabled bool `json:"dxr_enabled"`
	AVXEnabled bool `json:"avx_enabled"`
	DXVKAsync bool `json:"dxvk_async"`
	DXVKHUD string `json:"dxvk_hud"`
	DXVKFrameRate int `json:"dxvk_frame_rate"` // 0 = unlimited, otherwise caps FPS (DXVK only)
	RetinaMode bool `json:"retina_mode"`
	ResolutionScale string `json:"resolution_scale"`
	ManualOverride bool `json:"manual_override"` // Tells AI curation to back off
	Engine         string            `json:"engine"`
	ExecutablePath string            `json:"executable_path"`
	LaunchArgs     []string          `json:"launch_args"`
	DLLOverrides   map[string]string `json:"dll_overrides"`
	EnvVars        map[string]string `json:"env_vars"`
	Winetricks     []string          `json:"winetricks,omitempty"`
	Installed      bool              `json:"installed,omitempty"`
	ID             string            `json:"id,omitempty"`
	AppID          string            `json:"appID,omitempty"`
	Platform       string            `json:"platform,omitempty"`
	GameName       string            `json:"gameName,omitempty"`
}

// CloudGameConfig matches the expected JSON response from the Edge Function.
type CloudGameConfig struct {
	Engine         string            `json:"engine"`
	ExecutablePath string            `json:"executable_path"`
	LaunchArgs     []string          `json:"launch_args"`
	DLLOverrides   map[string]string `json:"dll_overrides"`
	EnvVars        map[string]string `json:"env_vars"`
	Winetricks     []string          `json:"winetricks"`
	WindowsVersion string            `json:"windows_version"`
	Confidence     float64           `json:"confidence"`
	AIProvider     string            `json:"ai_provider"`
	AIModel        string            `json:"ai_model"`
}
