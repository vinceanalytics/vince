import React, { createContext, PropsWithChildren, useContext, useEffect, useState } from "react"

import { VinceClient, createClient } from "../../vince";
import { useLocalStorage } from "../LocalStorageProvider";


type Props = {}

type ContextProps = {
    vince: VinceClient | undefined
}

const defaultValues = {
    vince: undefined,
}

export const VinceContext = createContext<ContextProps>(defaultValues)

export const VinceProvider = ({ children }: PropsWithChildren<Props>) => {
    const { authPayload } = useLocalStorage()
    const [vince, setClient] = useState<VinceClient | undefined>(undefined)
    useEffect(() => {
        if (authPayload !== "") {
            setClient(createClient({ token: authPayload }))
        }
    }, [authPayload, setClient])
    return (
        <VinceContext.Provider
            value={{
                vince,
            }}
        >
            {children}
        </VinceContext.Provider>
    )
}

export const useVince = () => {
    return useContext(VinceContext)
}