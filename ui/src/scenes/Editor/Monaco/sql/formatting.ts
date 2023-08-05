import { IRange } from "monaco-editor"
import * as monaco from "monaco-editor"
import { format } from "sql-formatter"


export const documentFormattingEditProvider = {
    provideDocumentFormattingEdits(
        model: monaco.editor.IModel,
        options: monaco.languages.FormattingOptions,
    ) {
        const formatted = format(model.getValue(), {
            language: "mysql",
            tabWidth: options.tabSize,
        })
        return [
            {
                range: model.getFullModelRange(),
                text: formatted,
            },
        ]
    },
}

export const documentRangeFormattingEditProvider = {
    provideDocumentRangeFormattingEdits(
        model: monaco.editor.IModel,
        range: IRange,
        options: monaco.languages.FormattingOptions,
    ) {
        const formatted = format(model.getValueInRange(range), {
            language: "mysql",
            tabWidth: options.tabSize,
        })
        return [
            {
                range,
                text: formatted,
            },
        ]
    },
}

