import React, { createContext, PropsWithChildren, useContext, useEffect, useState } from "react"

import {
    VinceClient, createVinceClient,
    SitesClient, createSitesClient
} from "../../vince";
import { useLocalStorage } from "../LocalStorageProvider";


type Props = {}

type ContextProps = {
    vince: VinceClient | undefined
    sitesClient: SitesClient | undefined
}

const defaultValues = {
    vince: undefined,
    sitesClient: undefined,
}

export const VinceContext = createContext<ContextProps>(defaultValues)

export const VinceProvider = ({ children }: PropsWithChildren<Props>) => {
    const { authPayload } = useLocalStorage()
    const [vince, setVinceClient] = useState<VinceClient | undefined>(undefined)
    const [sitesClient, setSitesClient] = useState<SitesClient | undefined>(undefined)
    useEffect(() => {
        if (authPayload !== "") {
            setVinceClient(createVinceClient({ token: authPayload }))
            setSitesClient(createSitesClient({ token: authPayload }))
        }
    }, [authPayload, setVinceClient, setSitesClient])
    return (
        <VinceContext.Provider
            value={{
                vince, sitesClient: sitesClient,
            }}
        >
            {children}
        </VinceContext.Provider>
    )
}

export const useVince = () => {
    return useContext(VinceContext)
}