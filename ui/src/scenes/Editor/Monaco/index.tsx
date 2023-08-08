import React, { useContext, useEffect, useRef, useState } from "react"
import type { BaseSyntheticEvent } from "react"
import Editor, { Monaco, loader } from "@monaco-editor/react"
import { editor } from "monaco-editor"
import type { IDisposable, IRange } from "monaco-editor"
import { VinceContext, useEditor } from "../../../providers"
import { usePreferences } from "./usePreferences"
import {
    appendQuery,
    getErrorRange,
    getQueryRequestFromEditor,
    getQueryRequestFromLastExecutedQuery,
    VINCELanguageName,
    setErrorMarker,
    clearModelMarkers,
    getQueryFromCursor,
    findMatches,
    AppendQueryOptions,
} from "./utils"
import type { Request } from "./utils"
import { PaneContent } from "../../../components"
import type { ErrorResult, Site } from "../../../vince"
import * as VINCE from "../../../vince"
import Loader from "../Loader"
import styled from "styled-components"
import { Text, themeGet } from "@primer/react";
import {
    conf as QuestDBLanguageConf,
    language as QuestDBLanguage,
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
    background: ${themeGet("fg.default")};
  }

  .cursorQueryDecoration {
    width: 0.2rem !important;
    background: ${themeGet("success.fg")};
    margin-left: 1.2rem;

    &.hasError {
      background: ${themeGet("danger.fg")};
    }
  }

  .cursorQueryGlyph {
    margin-left: 2rem;
    z-index: 1;
    cursor: pointer;

    &:after {
      content: "◃";
      font-size: 2.5rem;
      transform: rotate(180deg) scaleX(0.8);
      color: ${themeGet("success.fg")};
    }
  }

  .errorGlyph {
    margin-left: 2.5rem;
    margin-top: 0.5rem;
    z-index: 1;
    width: 0.75rem !important;
    height: 0.75rem !important;
    border-radius: 50%;
    background: ${themeGet("danger.fg")};
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
    const [running, setRunning] = useState<boolean>(false)
    const [tables, setTables] = useState<Site[]>([]);
    const [schemaCompletionHandle, setSchemaCompletionHandle] =
        useState<IDisposable>()
    const decorationsRef = useRef<string[]>([])
    const errorRef = useRef<ErrorResult | undefined>()
    const errorRangeRef = useRef<IRange | undefined>()

    const toggleRunning = (isRefresh: boolean = false) => {
        // dispatch(actions.query.toggleRunning(isRefresh))
    }

    const handleEditorBeforeMount = (monaco: Monaco) => {
        monaco.languages.register({ id: VINCELanguageName })

        monaco.languages.setMonarchTokensProvider(
            VINCELanguageName,
            QuestDBLanguage,
        )

        monaco.languages.setLanguageConfiguration(
            VINCELanguageName,
            QuestDBLanguageConf,
        )

        monaco.languages.registerCompletionItemProvider(
            VINCELanguageName,
            createQuestDBCompletionProvider(),
        )

        monaco.languages.registerDocumentFormattingEditProvider(
            VINCELanguageName,
            documentFormattingEditProvider,
        )

        monaco.languages.registerDocumentRangeFormattingEditProvider(
            VINCELanguageName,
            documentRangeFormattingEditProvider,
        )

        setSchemaCompletionHandle(
            monaco.languages.registerCompletionItemProvider(
                VINCELanguageName,
                createSchemaCompletionProvider(tables),
            ),
        )
    }

    const handleEditorClick = (e: BaseSyntheticEvent) => {
        if (e.target.classList.contains("cursorQueryGlyph")) {
            editorRef?.current?.focus()
            toggleRunning()
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
        monaco.editor.setTheme("dracula")

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
                    toggleRunning()
                },
            })

            editor.addAction({
                id: Command.CLEANUP_NOTIFICATIONS,
                label: "Clear all notifications",
                keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyK],
                run: () => {
                    // dispatch(actions.query.cleanupNotifications())
                },
            })

            editor.onDidChangeCursorPosition(() => {
                renderLineMarkings(monaco, editor)
            })
        }

        loadPreferences(editor)

        // Insert query, if one is found in the URL
        const params = new URLSearchParams(window.location.search)
        // Support multi-line queries (URL encoded)
        const query = params.get("query")
        const model = editor.getModel()
        if (query && model) {
            // Find if the query is already in the editor
            const matches = findMatches(model, query)
            if (matches && matches.length > 0) {
                editor.setSelection(matches[0].range)
                // otherwise, append the query
            } else {
                appendQuery(editor, query, { appendAt: "end" })
            }
        }

        const executeQuery = params.get("executeQuery")
        if (executeQuery) {
            toggleRunning()
        }
    }

    useEffect(() => {
        if (!running && request) {
            vince.abort()
            // dispatch(actions.query.stopRunning())
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
                    .queryRaw(request.query, { limit: "0,1000", explain: true })
                    .then((result) => {
                        setRequest(undefined)
                        errorRef.current = undefined
                        errorRangeRef.current = undefined
                        // dispatch(actions.query.stopRunning())
                        // dispatch(actions.query.setResult(result))

                        if (monacoRef?.current && editorRef?.current) {
                            renderLineMarkings(monacoRef.current, editorRef?.current)
                        }

                        if (result.type === VINCE.Type.DDL) {
                            // dispatch(
                            //     actions.query.addNotification({
                            //         content: (
                            //             <Text color="foreground"
                            //                 // TODO:(gernest)  ellipsis
                            //                 title={result.query}>
                            //                 {result.query}
                            //             </Text>
                            //         ),
                            //     }),
                            // )
                        }

                        if (result.type === VINCE.Type.DQL) {
                            setLastExecutedQuery(request.query)
                            // dispatch(
                            //     actions.query.addNotification({
                            //         jitCompiled: result.explain?.jitCompiled ?? false,
                            //         content: (
                            //             <QueryResult {...result.timings} rowCount={result.count} />
                            //         ),
                            //         sideContent: (
                            //             <Text color="fg.default"
                            //                 // TODO:(gernest)  ellipsis
                            //                 title={result.query}>
                            //                 {result.query}
                            //             </Text>
                            //         ),
                            //     }),
                            // )
                        }
                    })
                    .catch((error: ErrorResult) => {
                        errorRef.current = error
                        setRequest(undefined)
                        // dispatch(actions.query.stopRunning())
                        // dispatch(
                        //     actions.query.addNotification({
                        //         content: <Text color="red">{error.error}</Text>,
                        //         sideContent: (
                        //             <Text color="foreground"
                        //                 // TODO:(gernest)  ellipsis

                        //                 title={request.query}>
                        //                 {request.query}
                        //             </Text>
                        //         ),
                        //         type: "error",
                        //     }),
                        // )

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
                // dispatch(actions.query.stopRunning())
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
                    VINCELanguageName,
                    createSchemaCompletionProvider(tables),
                ),
            )
        }
    }, [tables, monacoRef, editorReady])

    return (
        <Content onClick={handleEditorClick}>
            <Editor
                beforeMount={handleEditorBeforeMount}
                defaultLanguage={VINCELanguageName}
                onMount={handleEditorDidMount}
                options={{
                    fixedOverflowWidgets: true,
                    fontSize: 14,
                    glyphMargin: true,
                    renderLineHighlight: "gutter",
                    minimap: {
                        enabled: false,
                    },
                    selectOnLineNumbers: false,
                    scrollBeyondLastLine: false,
                    tabSize: 2,
                }}
            />
            <Loader show={!!request || !tables} />
        </Content>
    )
}

export default MonacoEditor