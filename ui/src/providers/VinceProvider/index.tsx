import React, { createContext, PropsWithChildren, useContext, useEffect, useState } from "react"

import {
    ClusterClient, createClusterClient,
    EventsClient, createEventsClient,
    GoalsClient, createGoalsClient,
    QueryClient, createQueryClient,
    SitesClient, createSitesClient,
    SnippetsClient, createSnippetsClient,
    VinceClient, createVinceClient,
    createImportClient, ImportClient,
} from "../../vince";
import { useTokenSource } from "../TokenSourceProvider";


type Props = {}

type ContextProps = {
    clusterClient: ClusterClient | undefined
    eventsClient: EventsClient | undefined
    goalsClient: GoalsClient | undefined
    importClient: ImportClient | undefined
    queryClient: QueryClient | undefined
    sitesClient: SitesClient | undefined
    snippetsClient: SnippetsClient | undefined
    vince: VinceClient | undefined
}

const defaultValues = {
    clusterClient: undefined,
    eventsClient: undefined,
    goalsClient: undefined,
    importClient: undefined,
    queryClient: undefined,
    sitesClient: undefined,
    snippetsClient: undefined,
    vince: undefined,
}

export const VinceContext = createContext<ContextProps>(defaultValues)

export const VinceProvider = ({ children }: PropsWithChildren<Props>) => {
    const { tokenSource } = useTokenSource()
    const [vince, setVinceClient] = useState<VinceClient | undefined>(undefined)
    const [sitesClient, setSitesClient] = useState<SitesClient | undefined>(undefined)
    const [queryClient, setQueryClient] = useState<QueryClient | undefined>(undefined)
    const [snippetsClient, setSnippetsClient] = useState<SnippetsClient | undefined>(undefined)
    const [clusterClient, setClusterClient] = useState<ClusterClient | undefined>(undefined)
    const [eventsClient, setEventsClient] = useState<EventsClient | undefined>(undefined)
    const [importClient, setImportClient] = useState<ImportClient | undefined>(undefined)
    const [goalsClient, setGoalsClient] = useState<GoalsClient | undefined>(undefined)
    useEffect(() => {
        if (tokenSource !== undefined) {
            setVinceClient(createVinceClient(tokenSource))
            setSitesClient(createSitesClient(tokenSource))
            setQueryClient(createQueryClient(tokenSource))
            setSnippetsClient(createSnippetsClient(tokenSource))
            setClusterClient(createClusterClient(tokenSource))
            setEventsClient(createEventsClient(tokenSource))
            setImportClient(createImportClient(tokenSource))
            setGoalsClient(createGoalsClient(tokenSource))
        }
    }, [tokenSource, setVinceClient, setSitesClient,
        setQueryClient, setSnippetsClient, setClusterClient, setEventsClient])
    return (
        <VinceContext.Provider
            value={{
                clusterClient,
                eventsClient,
                goalsClient,
                importClient,
                queryClient,
                snippetsClient,
                vince, sitesClient,
            }}
        >
            {children}
        </VinceContext.Provider>
    )
}

export const useVince = () => {
    return useContext(VinceContext)
}