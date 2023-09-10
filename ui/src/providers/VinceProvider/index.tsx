import React, { createContext, PropsWithChildren, useContext, useEffect, useState } from "react"

import {
    VinceClient, createVinceClient,
    SitesClient, createSitesClient,
    QueryClient, createQueryClient,
    SnippetsClient, createSnippetsClient
} from "../../vince";
import { useLocalStorage } from "../LocalStorageProvider";
import { useTokenSource } from "../TokenSourceProvider";


type Props = {}

type ContextProps = {
    vince: VinceClient | undefined
    sitesClient: SitesClient | undefined
    queryClient: QueryClient | undefined
    snippetsClient: SnippetsClient | undefined
}

const defaultValues = {
    vince: undefined,
    sitesClient: undefined,
    queryClient: undefined,
    snippetsClient: undefined,
}

export const VinceContext = createContext<ContextProps>(defaultValues)

export const VinceProvider = ({ children }: PropsWithChildren<Props>) => {
    const { tokenSource } = useTokenSource()
    const [vince, setVinceClient] = useState<VinceClient | undefined>(undefined)
    const [sitesClient, setSitesClient] = useState<SitesClient | undefined>(undefined)
    const [queryClient, setQueryClient] = useState<QueryClient | undefined>(undefined)
    const [snippetsClient, setSnippetsClient] = useState<SnippetsClient | undefined>(undefined)
    useEffect(() => {
        if (tokenSource !== undefined) {
            setVinceClient(createVinceClient(tokenSource))
            setSitesClient(createSitesClient(tokenSource))
            setQueryClient(createQueryClient(tokenSource))
            setSnippetsClient(createSnippetsClient(tokenSource))
        }
    }, [tokenSource, setVinceClient, setSitesClient, setQueryClient, setSnippetsClient])
    return (
        <VinceContext.Provider
            value={{
                vince, sitesClient, queryClient, snippetsClient
            }}
        >
            {children}
        </VinceContext.Provider>
    )
}

export const useVince = () => {
    return useContext(VinceContext)
}