// import { trimedPath } from "../modules/utils";
// import { History, SmockerError } from "../modules/types";
// import axios from "axios";
// import { useQuery } from "react-query";

// const fetchHistory = async (session: string): Promise<History> => {
//   const { data } = await axios.get(`${trimedPath}/history`, {
//     params: { session },
//   });
//   return data;
// };

// export function useHistory(session: string) {
//   return useQuery<History, SmockerError>(["history"], () => fetchHistory(session));
// }
