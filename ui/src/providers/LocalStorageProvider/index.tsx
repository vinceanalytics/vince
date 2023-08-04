import React, {
    createContext,
    PropsWithChildren,
    useState,
    useContext,
} from "react"
import {
    SettingsType,
    LocalConfig,
    StoreKey,
    getValue,
    setValue,
    parseInteger,
} from './types'

type Props = {}


type ContextProps = {
    authPayload: string
    editorSplitterBasis: number
    resultsSplitterBasis: number
    updateSettings: (key: StoreKey, value: SettingsType) => void
}
const defaultConfig: LocalConfig = {
    authPayload: "",
    editorSplitterBasis: 350,
    resultsSplitterBasis: 350,
}

const defaultValues: ContextProps = {
    authPayload: "",
    editorSplitterBasis: 350,
    resultsSplitterBasis: 350,
    updateSettings: (key: StoreKey, value: SettingsType) => undefined,
}


export const LocalStorageContext = createContext<ContextProps>(defaultValues)


export const LocalStorageProvider = ({
    children,
}: PropsWithChildren<Props>) => {
    const [authPayload, setAuthPayload] = useState<string>(
        getValue(StoreKey.AUTH_PAYLOAD),
    )
    const updateSettings = (key: StoreKey, value: SettingsType) => {
        setValue(key, value.toString())
        refreshSettings(key)
    }
    const [editorSplitterBasis, seteditorSplitterBasis] = useState<number>(
        parseInteger(
            getValue(StoreKey.EDITOR_SPLITTER_BASIS),
            defaultConfig.editorSplitterBasis,
        ),
    )
    const [resultsSplitterBasis, setresultsSplitterBasis] = useState<number>(
        parseInteger(
            getValue(StoreKey.RESULTS_SPLITTER_BASIS),
            defaultConfig.resultsSplitterBasis,
        ),
    )

    const refreshSettings = (key: StoreKey) => {
        const value = getValue(key)
        switch (key) {
            case StoreKey.AUTH_PAYLOAD:
                setAuthPayload(value)
                break
            case StoreKey.EDITOR_SPLITTER_BASIS:
                seteditorSplitterBasis(
                    parseInteger(value, defaultConfig.editorSplitterBasis),
                )
                break
            case StoreKey.RESULTS_SPLITTER_BASIS:
                setresultsSplitterBasis(
                    parseInteger(value, defaultConfig.resultsSplitterBasis),
                )
                break
        }
    }
    return (
        <LocalStorageContext.Provider
            value={{
                authPayload,
                editorSplitterBasis,
                resultsSplitterBasis,
                updateSettings,
            }}
        >
            {children}
        </LocalStorageContext.Provider>
    )
}

export const useLocalStorage = () => {
    return useContext(LocalStorageContext)
}

export * from "./types"
