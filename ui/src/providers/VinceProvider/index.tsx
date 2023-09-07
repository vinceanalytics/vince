import React, { createContext, PropsWithChildren, useContext, useEffect, useState } from "react"

import {
    VinceClient, createVinceClient,
    SitesClient, createSitesClient,
    QueryClient, createQueryClient
} from "../../vince";
import { useLocalStorage } from "../LocalStorageProvider";


type Props = {}

type ContextProps = {
    vince: VinceClient | undefined
    sitesClient: SitesClient | undefined
    queryClient: QueryClient | undefined
}

const defaultValues = {
    vince: undefined,
    sitesClient: undefined,
    queryClient: undefined,
}

export const VinceContext = createContext<ContextProps>(defaultValues)

export const VinceProvider = ({ children }: PropsWithChildren<Props>) => {
    const { authPayload } = useLocalStorage()
    const [vince, setVinceClient] = useState<VinceClient | undefined>(undefined)
    const [sitesClient, setSitesClient] = useState<SitesClient | undefined>(undefined)
    const [queryClient, setQueryClient] = useState<QueryClient | undefined>(undefined)
    useEffect(() => {
        if (authPayload !== "") {
            setVinceClient(createVinceClient({ token: authPayload }))
            setSitesClient(createSitesClient({ token: authPayload }))
            setQueryClient(createQueryClient({ token: authPayload }))
        }
    }, [authPayload, setVinceClient, setSitesClient])
    return (
        <VinceContext.Provider
            value={{
                vince, sitesClient, queryClient,
            }}
        >
            {children}
        </VinceContext.Provider>
    )
}

export const useVince = () => {
    return useContext(VinceContext)
}