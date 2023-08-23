import { createContext, PropsWithChildren, useState, useContext, useEffect, useCallback } from "react"
import { QueryResult } from "../../vince"
import { useEditor } from "../EditorProvider"
import { getQueryRequestFromEditor } from "../../scenes/Editor/Monaco/utils"


type ContextProps = {
    running: boolean
    toggleRun: () => void
    stopRunning: () => void
}

const defaultValues = {
    running: false,
    toggleRun: () => undefined,
    stopRunning: () => undefined,
}

const QueryContext = createContext<ContextProps>(defaultValues)


export const QueryProvider = ({ children }: PropsWithChildren<{}>) => {
    const [running, setRunning] = useState<boolean>(false)

    const toggleRun = () => {
        setRunning(!running)
    }
    const stopRunning = () => {
    }

    return (
        <QueryContext.Provider
            value={{ running, toggleRun, stopRunning }}
        >
            {children}
        </QueryContext.Provider>
    )
}

export const useQuery = () => {
    return useContext(QueryContext)
}