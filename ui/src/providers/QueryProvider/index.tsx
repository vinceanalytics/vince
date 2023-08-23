import { createContext, PropsWithChildren, useState, useContext, useEffect, useCallback } from "react"
import { QueryResult } from "../../vince"
import { useEditor } from "../EditorProvider"
import { getQueryRequestFromEditor } from "../../scenes/Editor/Monaco/utils"


type ContextProps = {
    running: boolean
    request: string
    toggleRun: () => void
    stopRunning: () => void
}

const defaultValues = {
    running: false,
    request: "",
    toggleRun: () => undefined,
    stopRunning: () => undefined,
}

const QueryContext = createContext<ContextProps>(defaultValues)


export const QueryProvider = ({ children }: PropsWithChildren<{}>) => {
    const [running, setRunning] = useState<boolean>(false)
    const [request, setRequest] = useState<string>("")
    const { editorRef } = useEditor()

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
    }

    return (
        <QueryContext.Provider
            value={{ running, toggleRun, stopRunning, request }}
        >
            {children}
        </QueryContext.Provider>
    )
}

export const useQuery = () => {
    return useContext(QueryContext)
}