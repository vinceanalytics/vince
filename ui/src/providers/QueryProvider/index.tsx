import { createContext, PropsWithChildren, useState, useContext, useEffect, useCallback } from "react"
import { QueryResult, ErrorResult } from "../../vince"
import { useEditor } from "../EditorProvider"
import { getQueryRequestFromEditor } from "../../scenes/Editor/Monaco/utils"
import { useVince } from "../VinceProvider"


type ContextProps = {
    running: boolean
    request: string
    result: QueryResult
    toggleRun: () => void
    stopRunning: () => void
}

const emptyResult = {
    elapsed: "", columns: [], rows: []
}

const defaultValues = {
    running: false,
    request: "",
    result: emptyResult,
    toggleRun: () => undefined,
    stopRunning: () => undefined,
}

const QueryContext = createContext<ContextProps>(defaultValues)


export const QueryProvider = ({ children }: PropsWithChildren<{}>) => {
    const [running, setRunning] = useState<boolean>(false)
    const [request, setRequest] = useState<string>("")
    const [result, setResult] = useState<QueryResult>(emptyResult)
    const { editorRef } = useEditor()
    const { vince } = useVince()
    useEffect(() => {
        if (running && request !== "") {
            vince.query({ query: request })
                .then((result) => {
                    const q = result as QueryResult;
                    setResult(q)
                })
                .catch((error) => {
                    console.log(error)
                })
        }
    }, [running, vince, request])

    useEffect(() => {
        if (running && result) {
            setRunning(false)
        }
    }, [running, result])

    const toggleRun = () => {
        setRunning(!running)
        if (!running) {
            const query = getQueryRequestFromEditor(editorRef?.current!)
            if (query) {
                setRequest(query.query)
            }
        }
    }

    const stopRunning = () => {
        vince.abort()
        setRunning(false)
    }

    return (
        <QueryContext.Provider
            value={{ running, toggleRun, stopRunning, request, result }}
        >
            {children}
        </QueryContext.Provider>
    )
}

export const useQuery = () => {
    return useContext(QueryContext)
}