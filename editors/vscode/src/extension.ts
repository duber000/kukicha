import * as fs from 'fs';
import * as path from 'path';
import * as vscode from 'vscode';
import {
    LanguageClient,
    LanguageClientOptions,
    ServerOptions,
    TransportKind,
    Executable
} from 'vscode-languageclient/node';

let client: LanguageClient;

export function activate(context: vscode.ExtensionContext) {
    console.log('Kukicha extension is now active!');

    const config = vscode.workspace.getConfiguration('kukicha');
    const customLspPath = config.get<string>('lsp.path', '');

    const lspPath = customLspPath || getBundledLspPath(context);

    if (!lspPath) {
        vscode.window.showErrorMessage(
            'Kukicha LSP binary not found. Please install the kukicha compiler or set the path manually.'
        );
        return;
    }

    const serverExecutable: Executable = {
        command: lspPath,
        transport: TransportKind.stdio,
        options: {
            env: process.env
        }
    };

    const serverOptions: ServerOptions = {
        run: serverExecutable,
        debug: serverExecutable
    };

    const clientOptions: LanguageClientOptions = {
        documentSelector: [{ language: 'kukicha' }],
        synchronize: {
            configurationSection: 'kukicha',
            fileEvents: vscode.workspace.createFileSystemWatcher('**/*.kuki')
        },
        initializationOptions: {
            trace: config.get<string>('trace.server', 'off')
        },
        outputChannel: vscode.window.createOutputChannel('Kukicha Language Server')
    };

    client = new LanguageClient(
        'kukicha',
        'Kukicha Language Server',
        serverOptions,
        clientOptions
    );

    client.start();

    registerCommands(context);
    registerFormatOnSave(context);
}

export function deactivate(): Thenable<void> | undefined {
    if (!client) {
        return undefined;
    }
    return client.stop();
}

function getBundledLspPath(context: vscode.ExtensionContext): string {
    const platform = process.platform;
    const arch = process.arch;
    
    let platformDir: string;
    
    if (platform === 'win32') {
        platformDir = 'win32';
    } else if (platform === 'darwin') {
        platformDir = arch === 'arm64' ? 'darwin-arm64' : 'darwin-x64';
    } else {
        platformDir = 'linux';
    }

    const binaryName = platform === 'win32' ? 'kukicha-lsp.exe' : 'kukicha-lsp';
    const lspPath = path.join(
        context.extensionPath,
        'binaries',
        platformDir,
        binaryName
    );

    return fs.existsSync(lspPath) ? lspPath : '';
}

function registerCommands(context: vscode.ExtensionContext) {
    const formatCommand = vscode.commands.registerCommand('kukicha.format', async () => {
        const editor = vscode.window.activeTextEditor;
        if (!editor) {
            return;
        }

        const document = editor.document;
        if (document.languageId !== 'kukicha') {
            return;
        }

        try {
            await vscode.commands.executeCommand('editor.action.formatDocument');
        } catch (error) {
            vscode.window.showErrorMessage(`Failed to format: ${error}`);
        }
    });

    const checkSyntaxCommand = vscode.commands.registerCommand('kukicha.checkSyntax', async () => {
        const editor = vscode.window.activeTextEditor;
        if (!editor) {
            return;
        }

        const document = editor.document;
        if (document.languageId !== 'kukicha') {
            return;
        }

        try {
            await vscode.commands.executeCommand('workbench.action.tasks.runTask', 'Kukicha: Check Syntax');
        } catch (error) {
            vscode.window.showErrorMessage(`Syntax check failed: ${error}`);
        }
    });

    context.subscriptions.push(formatCommand, checkSyntaxCommand);
}

function registerFormatOnSave(context: vscode.ExtensionContext) {
    const saveDisposable = vscode.workspace.onWillSaveTextDocument((event) => {
        const config = vscode.workspace.getConfiguration('kukicha');
        if (!config.get<boolean>('format.enable', true)) {
            return;
        }
        if (event.document.languageId !== 'kukicha') {
            return;
        }
        event.waitUntil(
            vscode.commands.executeCommand<vscode.TextEdit[]>(
                'vscode.executeFormatDocumentProvider',
                event.document.uri,
                { tabSize: 4, insertSpaces: true }
            ).then(edits => edits ?? [])
        );
    });
    context.subscriptions.push(saveDisposable);
}
