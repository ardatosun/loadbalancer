import React, { useEffect, useState } from "react";
import { TbArrowsDownUp } from "react-icons/tb";

export default function Sort({ data, updateData }) {

    const [toggleSort, setToggleSort] = useState(false);
    const [sortOrder, setSortOrder] = useState("Ascending");
    const [sortType, setSortType] = useState("id");

    useEffect(()=>{
        updateData((prevSortedData) => {
            const sortedArray = [...data].sort((a, b) => {
                const aValue = String(a[sortType]).toLowerCase(); // Ensure string comparison
                const bValue = String(b[sortType]).toLowerCase(); // Ensure string comparison

                if (sortOrder === 'Ascending') {
                    return aValue.localeCompare(bValue); // Ascending order
                } else {
                    return bValue.localeCompare(aValue); // Descending order
                }
            });

            return sortedArray;
        });
    },[sortType, sortOrder])

    return (
        <div className="h-full relative">
            <div className="h-full p-2" onClick={() => setToggleSort(!toggleSort)}>
                <div className="border-2 border-limegreen rounded-md h-full w-full cursor-pointer bg-black hover:bg-grey px-3 gap-1.5 justify-center flex items-center">
                        <div><TbArrowsDownUp color="#62d5a4" size={15} /></div>
                        <span className="text-s tracking-wider">Sort</span>
                </div>
            </div>
            {toggleSort &&
            <div className="absolute w-56 h-32 bg-lightgrey right-0 rounded-lg flex flex-col gap-4 p-5">
                <select 
                    className="rounded-lg bg-grey h-8 px-4 border focus:border-limegreen text-sm tracking-wider"
                    onChange={(e)=>{setSortOrder(e.target.value)}}
                    value={sortOrder}
                >
                    <option>Ascending</option>
                    <option>Descending</option>
                </select>
                <select 
                    className="rounded-lg bg-grey h-8 px-4 border focus:border-limegreen text-sm tracking-wider"
                    onChange={(e)=>{setSortType(e.target.value)}}
                    value={sortType}
                >
                    {(data.length > 0 && Object.keys(data[0]).map((keys, index)=>{
                        return <option key={index} value={keys}>
                            {keys === "id" && "ID"}
                            {keys === "ip_address" && "IP Address"}
                            {keys === "status" && "Status"}
                            {keys === "last_checked" && "Last Checked"}
                        </option>
                    }))}
                </select>
            </div>
            }
            
        </div>
        
    )
}