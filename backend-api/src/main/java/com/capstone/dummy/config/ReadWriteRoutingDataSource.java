package com.capstone.dummy.config;

import org.springframework.jdbc.datasource.lookup.AbstractRoutingDataSource;
import org.springframework.transaction.support.TransactionSynchronizationManager;

public class ReadWriteRoutingDataSource extends AbstractRoutingDataSource {
    
    @Override
    protected Object determineCurrentLookupKey() {
        // If current transaction is readOnly context from @Transactional(readOnly = true), use "replica"
        return TransactionSynchronizationManager.isCurrentTransactionReadOnly() ? "replica" : "master";
    }
}
