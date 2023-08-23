import { Site } from "../../../../vince"
import * as monaco from "monaco-editor"
import { CompletionItemKind } from "./types"

export const createSchemaCompletionProvider = (questDBTables: Site[] = []) => {
    const completionProvider: monaco.languages.CompletionItemProvider = {
        triggerCharacters: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz \"".split(
            "",
        ),
        provideCompletionItems(model, position) {
            const word = model.getWordUntilPosition(position)

            const textUntilPosition = model.getValueInRange({
                startLineNumber: 1,
                startColumn: 1,
                endLineNumber: position.lineNumber,
                endColumn: word.startColumn,
            })

            const range = {
                startLineNumber: position.lineNumber,
                endLineNumber: position.lineNumber,
                startColumn: word.startColumn,
                endColumn: word.endColumn,
            }

            if (
                word.word ||
                /(FROM|INTO|TABLE) $/gim.test(textUntilPosition)
            ) {
                return {
                    suggestions: questDBTables.map((item) => {
                        return {
                            label: item.domain,
                            kind: CompletionItemKind.Class,
                            insertText: `\`${item.domain}\``,
                            range,
                        }
                    }),
                }
            }
        },
    }

    return completionProvider
}