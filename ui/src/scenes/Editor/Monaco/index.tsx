import React, { useContext, useEffect, useRef, useState } from "react"
import type { BaseSyntheticEvent } from "react"
import Editor, { Monaco, loader } from "@monaco-editor/react"
import { editor } from "monaco-editor"
import type { IDisposable, IRange } from "monaco-editor"
import { VinceContext, useEditor, useQuery, useSites } from "../../../providers"
import { usePreferences } from "./usePreferences"
import {
    appendQuery,
    getErrorRange,
    getQueryRequestFromEditor,
    getQueryRequestFromLastExecutedQuery,
    setErrorMarker,
    clearModelMarkers,
    getQueryFromCursor,
    findMatches,
} from "./utils"
import type { Request } from "./utils"
import { PaneContent } from "../../../components"
import type { ErrorResult } from "../../../vince"
import styled from "styled-components"

import {
    createQuestDBCompletionProvider,
    createSchemaCompletionProvider,
    documentFormattingEditProvider,
    documentRangeFormattingEditProvider,
} from "./sql"

loader.config({
    paths: {
        vs: "vs",
    },
})

type IStandaloneCodeEditor = editor.IStandaloneCodeEditor

const Content = styled(PaneContent)`
  position: relative;
  overflow: hidden;

  .monaco-scrollable-element > .scrollbar > .slider {
    background: #1f883d ;
  }

  .cursorQueryDecoration {
    width: 0.2rem !important;
    background: #1f883d;
    margin-left: 1.2rem;

    &.hasError {
      background: #1f883d;
    }
  }

  .cursorQueryGlyph {
    margin-left: 2rem;
    z-index: 1;
    cursor: pointer;

    &:after {
      content: "◃";
      transform: rotate(180deg) scaleX(0.8);
      color: #1f883d;
    }
  }

  .errorGlyph {
    margin-left: 2.5rem;
    margin-top: 0.5rem;
    z-index: 1;
    width: 0.75rem !important;
    height: 0.75rem !important;
    border-radius: 50%;
    background: #cf222e;
  }
`

enum Command {
    EXECUTE = "execute",
    FOCUS_GRID = "focus_grid",
    CLEANUP_NOTIFICATIONS = "clean_notifications",
}

const MonacoEditor = () => {
    const { editorRef, monacoRef, insertTextAtCursor } = useEditor()
    const { loadPreferences, savePreferences } = usePreferences()
    const { vince } = useContext(VinceContext)
    const [request, setRequest] = useState<Request | undefined>()
    const [editorReady, setEditorReady] = useState<boolean>(false)
    const [lastExecutedQuery, setLastExecutedQuery] = useState("")
    const { running, toggleRun, stopRunning } = useQuery()
    const { sites } = useSites()
    const [schemaCompletionHandle, setSchemaCompletionHandle] =
        useState<IDisposable>()
    const decorationsRef = useRef<string[]>([])
    const errorRef = useRef<ErrorResult | undefined>()
    const errorRangeRef = useRef<IRange | undefined>()

    const handleEditorBeforeMount = (monaco: Monaco) => {

        monaco.languages.registerCompletionItemProvider(
            "mysql",
            createQuestDBCompletionProvider(),
        )

        monaco.languages.registerDocumentFormattingEditProvider(
            "mysql",
            documentFormattingEditProvider,
        )

        monaco.languages.registerDocumentRangeFormattingEditProvider(
            "mysql",
            documentRangeFormattingEditProvider,
        )

        setSchemaCompletionHandle(
            monaco.languages.registerCompletionItemProvider(
                "mysql",
                createSchemaCompletionProvider(sites),
            ),
        )
    }

    const handleEditorClick = (e: BaseSyntheticEvent) => {
        if (e.target.classList.contains("cursorQueryGlyph")) {
            editorRef?.current?.focus()
            toggleRun()
        }
    }

    const renderLineMarkings = (
        monaco: Monaco,
        editor: IStandaloneCodeEditor,
    ) => {
        const queryAtCursor = getQueryFromCursor(editor)
        const model = editor.getModel()
        if (queryAtCursor && model !== null) {
            const matches = findMatches(model, queryAtCursor.query)

            if (matches.length > 0) {
                const hasError = errorRef.current?.query === queryAtCursor.query
                const cursorMatch = matches.find(
                    (m) => m.range.startLineNumber === queryAtCursor.row + 1,
                )
                if (cursorMatch) {
                    decorationsRef.current = editor.deltaDecorations(
                        decorationsRef.current,
                        [
                            {
                                range: new monaco.Range(
                                    cursorMatch.range.startLineNumber,
                                    1,
                                    cursorMatch.range.endLineNumber,
                                    1,
                                ),
                                options: {
                                    isWholeLine: true,
                                    linesDecorationsClassName: `cursorQueryDecoration ${hasError ? "hasError" : ""
                                        }`,
                                },
                            },
                            {
                                range: new monaco.Range(
                                    cursorMatch.range.startLineNumber,
                                    1,
                                    cursorMatch.range.startLineNumber,
                                    1,
                                ),
                                options: {
                                    isWholeLine: false,
                                    glyphMarginClassName: "cursorQueryGlyph",
                                },
                            },
                            ...(errorRangeRef.current &&
                                cursorMatch.range.startLineNumber !==
                                errorRangeRef.current.startLineNumber
                                ? [
                                    {
                                        range: new monaco.Range(
                                            errorRangeRef.current.startLineNumber,
                                            0,
                                            errorRangeRef.current.startLineNumber,
                                            0,
                                        ),
                                        options: {
                                            isWholeLine: false,
                                            glyphMarginClassName: "errorGlyph",
                                        },
                                    },
                                ]
                                : []),
                        ],
                    )
                }
            }
        }
    }

    const handleEditorDidMount = (
        editor: IStandaloneCodeEditor,
        monaco: Monaco,
    ) => {
        if (monacoRef) {
            monacoRef.current = monaco
            setEditorReady(true)
        }

        if (editorRef) {
            editorRef.current = editor
            editor.addAction({
                id: Command.FOCUS_GRID,
                label: "Focus Grid",
                keybindings: [monaco.KeyCode.F2],
                run: () => {
                },
            })

            editor.addAction({
                id: Command.EXECUTE,
                label: "Execute command",
                keybindings: [
                    monaco.KeyCode.F9,
                    monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter,
                ],
                run: () => {
                    toggleRun()
                },
            })

            editor.onDidChangeCursorPosition(() => {
                renderLineMarkings(monaco, editor)
            })
        }

        loadPreferences(editor)
    }

    useEffect(() => {
        if (!running && request) {
            vince.abort()
            stopRunning()
            setRequest(undefined)
        }
    }, [request, vince, running])

    useEffect(() => {
        if (running && editorRef?.current) {
            if (monacoRef?.current) {
                clearModelMarkers(monacoRef.current, editorRef.current)
            }

            const request = running
                ? getQueryRequestFromLastExecutedQuery(lastExecutedQuery)
                : getQueryRequestFromEditor(editorRef.current)

            if (request?.query) {
                void vince
                    .query({ query: request.query })
                    .then(() => {
                        setRequest(undefined)
                        errorRef.current = undefined
                        errorRangeRef.current = undefined
                        if (monacoRef?.current && editorRef?.current) {
                            renderLineMarkings(monacoRef.current, editorRef?.current)
                        }
                    })
                    .catch((error: ErrorResult) => {
                        errorRef.current = error
                        setRequest(undefined)
                        if (editorRef?.current && monacoRef?.current) {
                            const errorRange = getErrorRange(
                                editorRef.current,
                                request,
                                error.position,
                            )
                            errorRangeRef.current = errorRange ?? undefined
                            if (errorRange) {
                                setErrorMarker(
                                    monacoRef?.current,
                                    editorRef.current,
                                    errorRange,
                                    error.error,
                                )
                                renderLineMarkings(monacoRef?.current, editorRef?.current)
                            }
                        }
                    })
                setRequest(request)
            } else {
                stopRunning()
            }
        }
    }, [vince, running])

    useEffect(() => {
        const editor = editorRef?.current
        if (running && editor) {
            savePreferences(editor)
        }
    }, [running, savePreferences])

    useEffect(() => {
        if (editorReady && monacoRef?.current) {
            schemaCompletionHandle?.dispose()
            setSchemaCompletionHandle(
                monacoRef.current.languages.registerCompletionItemProvider(
                    "mysql",
                    createSchemaCompletionProvider(sites),
                ),
            )
        }
    }, [sites, monacoRef, editorReady])

    return (
        <Content onClick={handleEditorClick}>
            <Editor
                beforeMount={handleEditorBeforeMount}
                defaultLanguage={"mysql"}
                onMount={handleEditorDidMount}
                options={{
                    fixedOverflowWidgets: true,
                    fontSize: 14,
                    glyphMargin: true,
                    renderLineHighlight: "gutter",
                    minimap: {
                        enabled: false,
                    },
                    fontFamily: 'SFMono-Regular, Menlo, Monaco, Consolas,"Liberation Mono", "Courier New", monospace',
                    selectOnLineNumbers: false,
                    scrollBeyondLastLine: false,
                    tabSize: 2,
                }}
            />
        </Content>
    )
}

export default MonacoEditor