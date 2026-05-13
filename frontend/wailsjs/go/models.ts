export namespace ai {
	
	export class Usage {
	    prompt_tokens: number;
	    completion_tokens: number;
	    total_tokens: number;
	
	    static createFrom(source: any = {}) {
	        return new Usage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.prompt_tokens = source["prompt_tokens"];
	        this.completion_tokens = source["completion_tokens"];
	        this.total_tokens = source["total_tokens"];
	    }
	}
	export class ChatResponse {
	    id: string;
	    model: string;
	    content: string;
	    usage: Usage;
	
	    static createFrom(source: any = {}) {
	        return new ChatResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.model = source["model"];
	        this.content = source["content"];
	        this.usage = this.convertValues(source["usage"], Usage);
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
	export class Citation {
	    document_id: string;
	    document_name: string;
	    chunk_index: number;
	    snippet: string;
	    score: number;
	
	    static createFrom(source: any = {}) {
	        return new Citation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.document_id = source["document_id"];
	        this.document_name = source["document_name"];
	        this.chunk_index = source["chunk_index"];
	        this.snippet = source["snippet"];
	        this.score = source["score"];
	    }
	}
	export class Message {
	    role: string;
	    content: string;
	    display_content?: string;
	    citations?: Citation[];
	
	    static createFrom(source: any = {}) {
	        return new Message(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.role = source["role"];
	        this.content = source["content"];
	        this.display_content = source["display_content"];
	        this.citations = this.convertValues(source["citations"], Citation);
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
	export class ChatSession {
	    ID: string;
	    Title: string;
	    Provider: string;
	    Model: string;
	    Messages: Message[];
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    LastActiveAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ChatSession(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Title = source["Title"];
	        this.Provider = source["Provider"];
	        this.Model = source["Model"];
	        this.Messages = this.convertValues(source["Messages"], Message);
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.LastActiveAt = this.convertValues(source["LastActiveAt"], null);
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
	
	export class ConnectionTestResult {
	    success: boolean;
	    error?: string;
	    response_time?: number;
	    model?: string;
	    tokens_used?: number;
	    test_message?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	        this.response_time = source["response_time"];
	        this.model = source["model"];
	        this.tokens_used = source["tokens_used"];
	        this.test_message = source["test_message"];
	    }
	}
	
	export class ModelInfo {
	    id: string;
	    name: string;
	    max_tokens: number;
	    context_window: number;
	
	    static createFrom(source: any = {}) {
	        return new ModelInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.max_tokens = source["max_tokens"];
	        this.context_window = source["context_window"];
	    }
	}
	export class ProviderInfo {
	    name: string;
	    type: string;
	    enabled: boolean;
	    default_model: string;
	    models: ModelInfo[];
	
	    static createFrom(source: any = {}) {
	        return new ProviderInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.enabled = source["enabled"];
	        this.default_model = source["default_model"];
	        this.models = this.convertValues(source["models"], ModelInfo);
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

export namespace config {
	
	export class UIConfig {
	    SidebarCollapsed: boolean;
	    FontSize: number;
	    CodeTheme: string;
	
	    static createFrom(source: any = {}) {
	        return new UIConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.SidebarCollapsed = source["SidebarCollapsed"];
	        this.FontSize = source["FontSize"];
	        this.CodeTheme = source["CodeTheme"];
	    }
	}
	export class DefaultsConfig {
	    Workspace: string;
	    Workflow: string;
	
	    static createFrom(source: any = {}) {
	        return new DefaultsConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Workspace = source["Workspace"];
	        this.Workflow = source["Workflow"];
	    }
	}
	export class AppConfig {
	    Version: string;
	    Theme: string;
	    Language: string;
	    Defaults: DefaultsConfig;
	    UI: UIConfig;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Version = source["Version"];
	        this.Theme = source["Theme"];
	        this.Language = source["Language"];
	        this.Defaults = this.convertValues(source["Defaults"], DefaultsConfig);
	        this.UI = this.convertValues(source["UI"], UIConfig);
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
	export class PrivacyConfig {
	    Telemetry: boolean;
	    AutoUpdateCheck: boolean;
	    CrashReport: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PrivacyConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Telemetry = source["Telemetry"];
	        this.AutoUpdateCheck = source["AutoUpdateCheck"];
	        this.CrashReport = source["CrashReport"];
	    }
	}
	export class StorageConfig {
	    BaseDir: string;
	    MaxFileSize: number;
	    AllowedExts: string[];
	    LogAutoCleanup: boolean;
	    LogRetentionDays: number;
	
	    static createFrom(source: any = {}) {
	        return new StorageConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BaseDir = source["BaseDir"];
	        this.MaxFileSize = source["MaxFileSize"];
	        this.AllowedExts = source["AllowedExts"];
	        this.LogAutoCleanup = source["LogAutoCleanup"];
	        this.LogRetentionDays = source["LogRetentionDays"];
	    }
	}
	export class WorkflowConfig {
	    TemplateDirs: string[];
	    AutoUpdate: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WorkflowConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TemplateDirs = source["TemplateDirs"];
	        this.AutoUpdate = source["AutoUpdate"];
	    }
	}
	export class ModelConfig {
	    ID: string;
	    Name: string;
	    MaxTokens: number;
	    ContextWindow: number;
	
	    static createFrom(source: any = {}) {
	        return new ModelConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.MaxTokens = source["MaxTokens"];
	        this.ContextWindow = source["ContextWindow"];
	    }
	}
	export class ProviderConfig {
	    Name: string;
	    Type: string;
	    Endpoint: string;
	    APIKey: string;
	    DefaultModel: string;
	    Models: ModelConfig[];
	    Timeout: number;
	    Enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProviderConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Type = source["Type"];
	        this.Endpoint = source["Endpoint"];
	        this.APIKey = source["APIKey"];
	        this.DefaultModel = source["DefaultModel"];
	        this.Models = this.convertValues(source["Models"], ModelConfig);
	        this.Timeout = source["Timeout"];
	        this.Enabled = source["Enabled"];
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
	export class Config {
	    App: AppConfig;
	    Providers: ProviderConfig[];
	    Workflows: WorkflowConfig;
	    Storage: StorageConfig;
	    Privacy: PrivacyConfig;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.App = this.convertValues(source["App"], AppConfig);
	        this.Providers = this.convertValues(source["Providers"], ProviderConfig);
	        this.Workflows = this.convertValues(source["Workflows"], WorkflowConfig);
	        this.Storage = this.convertValues(source["Storage"], StorageConfig);
	        this.Privacy = this.convertValues(source["Privacy"], PrivacyConfig);
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

export namespace document {
	
	export class Document {
	    id: string;
	    name: string;
	    original_name: string;
	    type: string;
	    size: number;
	    text_content: string;
	    workspace: string;
	    tags: string[];
	    source: string;
	    // Go type: time
	    publish_date?: any;
	    issuer: string;
	    // Go type: time
	    effective_date?: any;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Document(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.original_name = source["original_name"];
	        this.type = source["type"];
	        this.size = source["size"];
	        this.text_content = source["text_content"];
	        this.workspace = source["workspace"];
	        this.tags = source["tags"];
	        this.source = source["source"];
	        this.publish_date = this.convertValues(source["publish_date"], null);
	        this.issuer = source["issuer"];
	        this.effective_date = this.convertValues(source["effective_date"], null);
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
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

export namespace openai {
	
	export class FunctionCall {
	    name?: string;
	    arguments?: string;
	
	    static createFrom(source: any = {}) {
	        return new FunctionCall(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.arguments = source["arguments"];
	    }
	}
	export class ToolCall {
	    index?: number;
	    id: string;
	    type: string;
	    function: FunctionCall;
	
	    static createFrom(source: any = {}) {
	        return new ToolCall(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.id = source["id"];
	        this.type = source["type"];
	        this.function = this.convertValues(source["function"], FunctionCall);
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

export namespace storage {
	
	export class Workspace {
	    ID: string;
	    name: string;
	    description: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Workspace(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.created_at = this.convertValues(source["created_at"], null);
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

export namespace workflow {
	
	export class ExecutionMetrics {
	    tokens_input: number;
	    tokens_output: number;
	    api_calls: number;
	    cost_usd: number;
	    duration: number;
	
	    static createFrom(source: any = {}) {
	        return new ExecutionMetrics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tokens_input = source["tokens_input"];
	        this.tokens_output = source["tokens_output"];
	        this.api_calls = source["api_calls"];
	        this.cost_usd = source["cost_usd"];
	        this.duration = source["duration"];
	    }
	}
	export class ExecutionRecord {
	    id: string;
	    workflow_id: string;
	    workflow_name: string;
	    status: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    started_at?: any;
	    // Go type: time
	    completed_at?: any;
	    summary: string;
	
	    static createFrom(source: any = {}) {
	        return new ExecutionRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workflow_id = source["workflow_id"];
	        this.workflow_name = source["workflow_name"];
	        this.status = source["status"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.started_at = this.convertValues(source["started_at"], null);
	        this.completed_at = this.convertValues(source["completed_at"], null);
	        this.summary = source["summary"];
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
	export class StepResult {
	    id: string;
	    name: string;
	    status: string;
	    output: Record<string, any>;
	    error?: string;
	    duration: number;
	
	    static createFrom(source: any = {}) {
	        return new StepResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.status = source["status"];
	        this.output = source["output"];
	        this.error = source["error"];
	        this.duration = source["duration"];
	    }
	}
	export class ExecutionResult {
	    id: string;
	    workflow_id: string;
	    status: string;
	    steps?: StepResult[];
	    outputs: Record<string, any>;
	    step_results: Record<string, StepResult>;
	    summary: string;
	    metrics: ExecutionMetrics;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    completed_at: any;
	
	    static createFrom(source: any = {}) {
	        return new ExecutionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workflow_id = source["workflow_id"];
	        this.status = source["status"];
	        this.steps = this.convertValues(source["steps"], StepResult);
	        this.outputs = source["outputs"];
	        this.step_results = this.convertValues(source["step_results"], StepResult, true);
	        this.summary = source["summary"];
	        this.metrics = this.convertValues(source["metrics"], ExecutionMetrics);
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.completed_at = this.convertValues(source["completed_at"], null);
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
	export class InputDef {
	    name: string;
	    type: string;
	    required: boolean;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new InputDef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.required = source["required"];
	        this.description = source["description"];
	    }
	}
	
	export class WorkflowInfo {
	    id: string;
	    name: string;
	    description: string;
	    version: string;
	    inputs: InputDef[];
	    steps: number;
	
	    static createFrom(source: any = {}) {
	        return new WorkflowInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.version = source["version"];
	        this.inputs = this.convertValues(source["inputs"], InputDef);
	        this.steps = source["steps"];
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
	export class WorkflowEditorData {
	    info: WorkflowInfo;
	    content: string;
	    source: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkflowEditorData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.info = this.convertValues(source["info"], WorkflowInfo);
	        this.content = source["content"];
	        this.source = source["source"];
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

