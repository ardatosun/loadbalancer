import React, { useEffect, useState } from "react";
import { MdOutlineSearch } from "react-icons/md";


export default function Searchbar({ data, updateData }) {

    const [searchText, setSearchText] = useState("");

    useEffect(() => {
        const trimmedSearchText = searchText.trim().toLowerCase();
        updateData(prevState => {
            if (trimmedSearchText !== "") {
                const updatedData = data
                    .filter(item =>
                        Object.values(item).some(value => 
                            String(value).toLowerCase().includes(trimmedSearchText)
                        )
                    )
                    .sort((a, b) => {
                        // Check for exact matches
                        const aMatch = Object.values(a).some(value => 
                            String(value).toLowerCase() === trimmedSearchText
                        );
                        const bMatch = Object.values(b).some(value => 
                            String(value).toLowerCase() === trimmedSearchText
                        );
                        // Prioritize exact matches
                        return bMatch - aMatch;
                    });
    
                return updatedData;
            } 
            return data;
        });
    }, [searchText]);

    return (
        <div className="w-11/12 p-4 h-full bg-black rounded-md flex">
           <div className=" w-6 flex items-center"><MdOutlineSearch color="#62d5a4" size={20}/></div>
           <input 
                className="w-full bg-black focus:outline-none tracking-wider"
                type="text"
                placeholder="Search..."
                onChange={(e)=>{setSearchText(e.target.value);}}
           />
        </div>
    )
}