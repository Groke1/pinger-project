import React from "react";
import Row from "./Row";
import TableHead from "./TableHead";

function Table({rows}) {
    if (!Array.isArray(rows) || rows.length === 0)  {
        return <TableHead/>;
    }
    return <table className="table table-striped table-bordered table-hover">
        <TableHead/>
        <tbody className="text-center table-group-divider">
            {rows.map((row, index) => (
                <Row key={index} {...row}/>
            )) }
        </tbody>
    </table>
}

export default Table;