import React, { Component } from 'react';
import TransactionsTable from './transactionTable';
import classes from './table.css';
import axios from "axios/index";
import ErrorMessage from './errorMessage';

class TransactionTable extends Component {
    constructor(props) {
        super(props);
        this.state = {
            data: []
        };
    }

    async componentDidMount() {
        try {
            const response = await axios.get(
                "http://localhost:8080/api/get_all_transactions")
            await this.setState({data: response.data});
        } catch (err) {
            console.log(err);
        }
    }

    render() {
        const table = this.state.data.map((data, i) => {
            const conversion = data.Cost / 10000000000000000000;
            return <TransactionsTable
                key={`${data.TxHash}${i}`}
                age={data.Age}
                txHash={data.TxHash}
                blockNumber={data.BlockNumber}
                to={data.To}
                from={data.From}
                value={data.Amount}
                cost={conversion}
                getBlockTransactions={this.props.getBlockTransactions}
                detailTransactionHandler={this.props.detailTransactionHandler}
                detailAccountHandler={this.props.detailAccountHandler}
            />
        })

        let combinedClasses = ['responsive-table', classes.table];
        return (

            <div>     
            {
                this.state.data.length > 0 ?  
                    <table key={this.state.data.TxHash} className={combinedClasses.join(' ')}>
                        <thead>
                            <tr>
                                <th scope="col" className={classes.thItem}> TxHash </th>
                                <th scope="col" className={classes.thItem}> Block </th>
                                <th scope="col" className={classes.thItem}> Age </th>
                                <th scope="col" className={classes.thItem}> From </th>                      
                                <th scope="col" className={classes.thItem}> To </th>
                                <th scope="col" className={classes.thItem}> Value </th>
                                <th scope="col" className={classes.thItem}> TxFee </th>
                            </tr>
                        </thead>
                        {table}
                    </table>
                : <ErrorMessage />
            } 
        </div>           
        );
    }
}
export default TransactionTable;
