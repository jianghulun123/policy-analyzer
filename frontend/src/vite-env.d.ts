/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

// Wails runtime types
declare global {
  interface Window {
    go: {
      main: {
        App: {
          GetVersion(): Promise<string>
          GetConfig(): Promise<any>
          UpdateConfig(config: any): Promise<void>
          GetWorkspaces(): Promise<any[]>
          CreateWorkspace(name: string, description: string): Promise<any>
          GetDocuments(workspaceID: string): Promise<any[]>
          ImportDocument(filePath: string, workspaceID: string, tags: string[]): Promise<any>
          GetDocumentContent(docID: string): Promise<string>
          DeleteDocument(docID: string): Promise<void>
          GetProviders(): Promise<any[]>
          GetModels(providerName: string): Promise<any[]>
          GetWorkflows(): Promise<any[]>
          ExecuteWorkflow(workflowID: string, inputs: any): Promise<string>
          CancelExecution(execID: string): Promise<void>
          GetAnalysisHistory(workspaceID: string): Promise<any[]>
          GetAnalysisResult(execID: string): Promise<any>
          ExportReport(execID: string, format: string): Promise<string>
        }
      }
    }
    runtime: {
      EventsOn(event: string, callback: (data: any) => void): void
      EventsOff(event: string): void
      EventsEmit(event: string, data?: any): void
    }
  }
}

export {}