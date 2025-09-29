CREATE OR REPLACE PROPERTY GRAPH DefaultGraph
  NODE TABLES (
    Accounts
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    AccountTransferAccounts
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Companies
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Partnerships
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Persons
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Products
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    ServiceRecords
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Vehicles
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS
  )
  EDGE TABLES (
    PersonOwnAccounts
      SOURCE KEY (person_id) REFERENCES Accounts (id)
      DESTINATION KEY (account_id) REFERENCES Persons (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    AccountTransfers
      SOURCE KEY (account_id) REFERENCES Accounts (id)
      DESTINATION KEY (transfer_id) REFERENCES Accounts (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    AccountTransferAccounts
      SOURCE KEY (id) REFERENCES AccountTransferAccounts (id)
      DESTINATION KEY (from_account_id) REFERENCES Accounts (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    AccountTransferAccounts
      SOURCE KEY (id) REFERENCES AccountTransferAccounts (id)
      DESTINATION KEY (to_account_id) REFERENCES Accounts (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Persons
      SOURCE KEY (company_employees) REFERENCES Companies (id)
      DESTINATION KEY (id) REFERENCES Persons (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Persons
      SOURCE KEY (company_ceo) REFERENCES Companies (id)
      DESTINATION KEY (id) REFERENCES Persons (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Partners
      SOURCE KEY (company_id) REFERENCES Companies (id)
      DESTINATION KEY (partner_id) REFERENCES Companies (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    company_subsidiaries
      SOURCE KEY (company_id) REFERENCES Companies (id)
      DESTINATION KEY (parent_id) REFERENCES Companies (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    company_subsidiaries
      SOURCE KEY (company_id) REFERENCES Companies (id)
      DESTINATION KEY (parent_id) REFERENCES Companies (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Products
      SOURCE KEY (company_products) REFERENCES Companies (id)
      DESTINATION KEY (id) REFERENCES Products (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Vehicles
      SOURCE KEY (company_manufactured_vehicles) REFERENCES Companies (id)
      DESTINATION KEY (id) REFERENCES Vehicles (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    ServiceRecords
      SOURCE KEY (company_service_records) REFERENCES Companies (id)
      DESTINATION KEY (id) REFERENCES ServiceRecords (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Partnerships
      SOURCE KEY (id) REFERENCES Partnerships (id)
      DESTINATION KEY (company_a_id) REFERENCES Companies (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Partnerships
      SOURCE KEY (id) REFERENCES Partnerships (id)
      DESTINATION KEY (company_b_id) REFERENCES Companies (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    PersonOwnAccounts
      SOURCE KEY (person_id) REFERENCES Persons (id)
      DESTINATION KEY (account_id) REFERENCES Accounts (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Persons
      SOURCE KEY (person_spouse) REFERENCES Persons (id)
      DESTINATION KEY (id) REFERENCES Persons (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Persons
      SOURCE KEY (id) REFERENCES Persons (id)
      DESTINATION KEY (person_spouse) REFERENCES Persons (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Friends
      SOURCE KEY (person_id) REFERENCES Persons (id)
      DESTINATION KEY (friend_id) REFERENCES Persons (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Persons
      SOURCE KEY (id) REFERENCES Persons (id)
      DESTINATION KEY (company_employees) REFERENCES Companies (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Persons
      SOURCE KEY (company_ceo) REFERENCES Persons (id)
      DESTINATION KEY (id) REFERENCES Companies (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Purchases
      SOURCE KEY (person_id) REFERENCES Persons (id)
      DESTINATION KEY (product_id) REFERENCES Products (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Vehicles
      SOURCE KEY (person_vehicles) REFERENCES Persons (id)
      DESTINATION KEY (id) REFERENCES Vehicles (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Products
      SOURCE KEY (id) REFERENCES Products (id)
      DESTINATION KEY (company_products) REFERENCES Companies (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Purchases
      SOURCE KEY (person_id) REFERENCES Products (id)
      DESTINATION KEY (product_id) REFERENCES Persons (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    product_recommendations
      SOURCE KEY (product_id) REFERENCES Products (id)
      DESTINATION KEY (recommended_by_id) REFERENCES Products (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    product_recommendations
      SOURCE KEY (product_id) REFERENCES Products (id)
      DESTINATION KEY (recommended_by_id) REFERENCES Products (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    ServiceRecords
      SOURCE KEY (id) REFERENCES ServiceRecords (id)
      DESTINATION KEY (vehicle_service_history) REFERENCES Vehicles (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    ServiceRecords
      SOURCE KEY (id) REFERENCES ServiceRecords (id)
      DESTINATION KEY (company_service_records) REFERENCES Companies (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Vehicles
      SOURCE KEY (id) REFERENCES Vehicles (id)
      DESTINATION KEY (person_vehicles) REFERENCES Persons (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    Vehicles
      SOURCE KEY (id) REFERENCES Vehicles (id)
      DESTINATION KEY (company_manufactured_vehicles) REFERENCES Companies (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS,
    ServiceRecords
      SOURCE KEY (vehicle_service_history) REFERENCES Vehicles (id)
      DESTINATION KEY (id) REFERENCES ServiceRecords (id)
      DEFAULT LABEL
      PROPERTIES ARE ALL COLUMNS
  );