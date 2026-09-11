export namespace adapters {
	
	export class RIFTUIProfile {
	    id: string;
	    name: string;
	    cover: string;
	    heroCover: string;
	    platform: string;
	    appID: string;
	    status: string;
	    playtime: string;
	    lastPlayed: string;
	    lastPlayedTs: number;
	    backend: string;
	    description: string;
	    tags: string[];
	    isInstalled: boolean;
	    isMacNative: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RIFTUIProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.cover = source["cover"];
	        this.heroCover = source["heroCover"];
	        this.platform = source["platform"];
	        this.appID = source["appID"];
	        this.status = source["status"];
	        this.playtime = source["playtime"];
	        this.lastPlayed = source["lastPlayed"];
	        this.lastPlayedTs = source["lastPlayedTs"];
	        this.backend = source["backend"];
	        this.description = source["description"];
	        this.tags = source["tags"];
	        this.isInstalled = source["isInstalled"];
	        this.isMacNative = source["isMacNative"];
	    }
	}

}

export namespace exec {
	
	export class Cmd {
	    Path: string;
	    Args: string[];
	    Env: string[];
	    Dir: string;
	    Stdin: any;
	    Stdout: any;
	    Stderr: any;
	    ExtraFiles: os.File[];
	    SysProcAttr?: syscall.SysProcAttr;
	    Process?: os.Process;
	    // Go type: os
	    ProcessState?: any;
	    Err: any;
	    WaitDelay: number;
	
	    static createFrom(source: any = {}) {
	        return new Cmd(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Args = source["Args"];
	        this.Env = source["Env"];
	        this.Dir = source["Dir"];
	        this.Stdin = source["Stdin"];
	        this.Stdout = source["Stdout"];
	        this.Stderr = source["Stderr"];
	        this.ExtraFiles = this.convertValues(source["ExtraFiles"], os.File);
	        this.SysProcAttr = this.convertValues(source["SysProcAttr"], syscall.SysProcAttr);
	        this.Process = this.convertValues(source["Process"], os.Process);
	        this.ProcessState = this.convertValues(source["ProcessState"], null);
	        this.Err = source["Err"];
	        this.WaitDelay = source["WaitDelay"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace os {
	
	export class Process {
	    Pid: number;
	
	    static createFrom(source: any = {}) {
	        return new Process(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Pid = source["Pid"];
	    }
	}

}

export namespace rift {
	
	export class ActiveDownloadRecord {
	    gameId: string;
	    platform: string;
	    appID: string;
	    gameName: string;
	    percent: number;
	    speed: number;
	    stage: string;
	    status: string;
	    startedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new ActiveDownloadRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.gameId = source["gameId"];
	        this.platform = source["platform"];
	        this.appID = source["appID"];
	        this.gameName = source["gameName"];
	        this.percent = source["percent"];
	        this.speed = source["speed"];
	        this.stage = source["stage"];
	        this.status = source["status"];
	        this.startedAt = source["startedAt"];
	    }
	}
	export class GameOverride {
	    engine: string;
	    retina: string;
	    hud: string;
	    args: string;
	    vram_limit: number;
	    fps_cap: number;
	    pre_purge: boolean;
	    game_mode: boolean;
	    executable_path: string;
	
	    static createFrom(source: any = {}) {
	        return new GameOverride(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.engine = source["engine"];
	        this.retina = source["retina"];
	        this.hud = source["hud"];
	        this.args = source["args"];
	        this.vram_limit = source["vram_limit"];
	        this.fps_cap = source["fps_cap"];
	        this.pre_purge = source["pre_purge"];
	        this.game_mode = source["game_mode"];
	        this.executable_path = source["executable_path"];
	    }
	}
	export class AppConfig {
	    library_folders: string[];
	    default_library: number;
	    library_dir?: string;
	    imported_games: Record<string, string>;
	    global_hud: boolean;
	    global_retina: boolean;
	    global_esync: boolean;
	    global_safe_mode: boolean;
	    disable_telemetry: boolean;
	    game_overrides: Record<string, GameOverride>;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.library_folders = source["library_folders"];
	        this.default_library = source["default_library"];
	        this.library_dir = source["library_dir"];
	        this.imported_games = source["imported_games"];
	        this.global_hud = source["global_hud"];
	        this.global_retina = source["global_retina"];
	        this.global_esync = source["global_esync"];
	        this.global_safe_mode = source["global_safe_mode"];
	        this.disable_telemetry = source["disable_telemetry"];
	        this.game_overrides = this.convertValues(source["game_overrides"], GameOverride, true);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AuthStatus {
	    epicConnected: boolean;
	    epicUsername: string;
	    steamConnected: boolean;
	    steamUsername: string;
	    steamInstalled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AuthStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.epicConnected = source["epicConnected"];
	        this.epicUsername = source["epicUsername"];
	        this.steamConnected = source["steamConnected"];
	        this.steamUsername = source["steamUsername"];
	        this.steamInstalled = source["steamInstalled"];
	    }
	}
	export class GameExecutableInfo {
	    relativePath: string;
	    fileName: string;
	    sizeBytes: number;
	    isSelected: boolean;
	    engine: string;
	    directXAPI: string;
	    is32Bit: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GameExecutableInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.relativePath = source["relativePath"];
	        this.fileName = source["fileName"];
	        this.sizeBytes = source["sizeBytes"];
	        this.isSelected = source["isSelected"];
	        this.engine = source["engine"];
	        this.directXAPI = source["directXAPI"];
	        this.is32Bit = source["is32Bit"];
	    }
	}
	
	export class StorageStats {
	    totalGamesSize: number;
	    totalCacheSize: number;
	    totalEngineSize: number;
	
	    static createFrom(source: any = {}) {
	        return new StorageStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalGamesSize = source["totalGamesSize"];
	        this.totalCacheSize = source["totalCacheSize"];
	        this.totalEngineSize = source["totalEngineSize"];
	    }
	}
	export class UpdateInfo {
	    updateAvailable: boolean;
	    currentVersion: string;
	    latestVersion: string;
	    releaseNotes: string;
	    downloadUrl: string;
	    publishedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.updateAvailable = source["updateAvailable"];
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.releaseNotes = source["releaseNotes"];
	        this.downloadUrl = source["downloadUrl"];
	        this.publishedAt = source["publishedAt"];
	    }
	}

}

export namespace runners {
	
	export class SelectivePack {
	    tag: string;
	    name: string;
	    description: string;
	    isRequired: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SelectivePack(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tag = source["tag"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.isRequired = source["isRequired"];
	    }
	}

}

export namespace syscall {
	
	export class Credential {
	    Uid: number;
	    Gid: number;
	    Groups: number[];
	    NoSetGroups: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Credential(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Uid = source["Uid"];
	        this.Gid = source["Gid"];
	        this.Groups = source["Groups"];
	        this.NoSetGroups = source["NoSetGroups"];
	    }
	}
	export class SysProcAttr {
	    Chroot: string;
	    Credential?: Credential;
	    Ptrace: boolean;
	    Setsid: boolean;
	    Setpgid: boolean;
	    Setctty: boolean;
	    Noctty: boolean;
	    Ctty: number;
	    Foreground: boolean;
	    Pgid: number;
	
	    static createFrom(source: any = {}) {
	        return new SysProcAttr(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Chroot = source["Chroot"];
	        this.Credential = this.convertValues(source["Credential"], Credential);
	        this.Ptrace = source["Ptrace"];
	        this.Setsid = source["Setsid"];
	        this.Setpgid = source["Setpgid"];
	        this.Setctty = source["Setctty"];
	        this.Noctty = source["Noctty"];
	        this.Ctty = source["Ctty"];
	        this.Foreground = source["Foreground"];
	        this.Pgid = source["Pgid"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace types {
	
	export class RuntimeConfig {
	    enhanced_sync: string;
	    metal_hud: boolean;
	    metal_trace: boolean;
	    dxr_enabled: boolean;
	    avx_enabled: boolean;
	    dxvk_async: boolean;
	    dxvk_hud: string;
	    dxvk_frame_rate: number;
	    retina_mode: boolean;
	    resolution_scale: string;
	    manual_override: boolean;
	    engine: string;
	    executable_path: string;
	    launch_args: string[];
	    dll_overrides: Record<string, string>;
	    env_vars: Record<string, string>;
	    winetricks?: string[];
	    installed?: boolean;
	    id?: string;
	    appID?: string;
	    platform?: string;
	    gameName?: string;
	
	    static createFrom(source: any = {}) {
	        return new RuntimeConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enhanced_sync = source["enhanced_sync"];
	        this.metal_hud = source["metal_hud"];
	        this.metal_trace = source["metal_trace"];
	        this.dxr_enabled = source["dxr_enabled"];
	        this.avx_enabled = source["avx_enabled"];
	        this.dxvk_async = source["dxvk_async"];
	        this.dxvk_hud = source["dxvk_hud"];
	        this.dxvk_frame_rate = source["dxvk_frame_rate"];
	        this.retina_mode = source["retina_mode"];
	        this.resolution_scale = source["resolution_scale"];
	        this.manual_override = source["manual_override"];
	        this.engine = source["engine"];
	        this.executable_path = source["executable_path"];
	        this.launch_args = source["launch_args"];
	        this.dll_overrides = source["dll_overrides"];
	        this.env_vars = source["env_vars"];
	        this.winetricks = source["winetricks"];
	        this.installed = source["installed"];
	        this.id = source["id"];
	        this.appID = source["appID"];
	        this.platform = source["platform"];
	        this.gameName = source["gameName"];
	    }
	}

}

