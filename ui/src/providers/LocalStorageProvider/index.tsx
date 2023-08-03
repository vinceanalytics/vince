import React, {
    createContext,
    PropsWithChildren,
    useState,
    useContext,
} from "react"
import {
    SettingsType,
    StoreKey, getValue, setValue,
} from './types'

type Props = {}


type ContextProps = {
    authPayload: string
    updateSettings: (key: StoreKey, value: SettingsType) => void
}

const defaultValues: ContextProps = {
    authPayload: "",
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

    const refreshSettings = (key: StoreKey) => {
        const value = getValue(key)
        switch (key) {
            case StoreKey.AUTH_PAYLOAD:
                setAuthPayload(value)
                break
        }
    }
    return (
        <LocalStorageContext.Provider
            value={{
                authPayload,
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


