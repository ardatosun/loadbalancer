import React, { useEffect, useState } from 'react';
import Chips from './Chips';
import moment from 'moment';

const DataTable = ({ data }) => {

  return (
    <table className="min-w-full border-collapse border border-lightgrey">
      <thead>
        <tr>
          <th className="border border-lightgrey p-2 tracking-wider">ID</th>
          <th className="border border-lightgrey p-2 tracking-wider">IP Address</th>
          <th className="border border-lightgrey p-2 tracking-wider">Status</th>
          <th className="border border-lightgrey p-2 tracking-wider">Last Checked</th>
        </tr>
      </thead>
      <tbody>
        {data.length > 0 && data.map((item) => {
          const id = item.id;
          const ip_address = item.ip_address;
          const status = item.status;
          const last_checked = item.last_checked;
          const formatted_date = moment(last_checked).format("MMMM D, YYYY h:mm A");
          return (
            <tr key={item.id} className="">
              <td className="border border-lightgrey py-2 px-4 tracking-wider"><span className='flex justify-center'>{id}</span></td>
              <td className="border border-lightgrey py-2 px-4 tracking-wider"><span className='flex justify-center'>{ip_address}</span></td>
              <td className="border border-lightgrey py-2 px-4 flex justify-center">
                  <Chips 
                    txtColor={status === "healthy" ? "text-limegreen" : "text-brightred"}
                    bgColor={status === "healthy" ? "bg-darkgreen" : "bg-darkred"}
                    ledColor={status === "healthy" ? "bg-limegreen" : "bg-brightred"}
                    text={status}
                  />
              </td>
              <td className="border border-lightgrey py-2 px-4 tracking-wider"><span className='flex justify-center'>{formatted_date}</span></td>
            </tr>
          )
        })}
      </tbody>
    </table>
  );
};

export default DataTable;
