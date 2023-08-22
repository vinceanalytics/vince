import React, { createContext, PropsWithChildren, useContext, useEffect, useState } from "react"

import { Client } from "../../vince";
import { useLocalStorage } from "../LocalStorageProvider";

const client = new Client();

type Props = {}

type ContextProps = {
    vince: Client
}

const defaultValues = {
    vince: client,
}

export const VinceContext = createContext<ContextProps>(defaultValues)

export const VinceProvider = ({ children }: PropsWithChildren<Props>) => {
    const { authPayload } = useLocalStorage()
    const [vince, setClient] = useState<Client>(client)
    useEffect(() => {
        if (authPayload !== "") {
            setClient(new Client(authPayload))
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