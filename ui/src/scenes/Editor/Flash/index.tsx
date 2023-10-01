import { Flash } from "@primer/react"
import { useQuery } from "../../../providers"

export const Error = () => {
    const { error } = useQuery()
    return (
        <>
            {error && <Flash full variant="danger">
                {error.message}
            </Flash>}
        </>
    )
}