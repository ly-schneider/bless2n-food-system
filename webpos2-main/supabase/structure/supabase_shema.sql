SELECT *
FROM
  information_schema.columns
where
table_name in ('company','discont','event','item','item_category','location','menu_stock','menu_tree','menuitems','order','order_fulfillment','order_line','order_status','payment_method','profile','role','setting','station','station_item')
ORDER BY
table_name,
ordinal_position;