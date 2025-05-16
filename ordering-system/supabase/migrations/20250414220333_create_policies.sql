CREATE POLICY "Admins can view their own data" 
  ON admins FOR SELECT USING (auth.uid() = id);

-- Products policies:
-- Public view: Allows anyone to view product details
CREATE POLICY "Public can view all products" 
  ON products FOR SELECT USING (TRUE);
-- Admins manage products:
CREATE POLICY "Admins can manage products" 
  ON products FOR ALL USING (
    EXISTS (SELECT 1 FROM admins WHERE id = auth.uid())
  );

-- Orders policies: Restrict orders only to admins
CREATE POLICY "Admins can view all orders" 
  ON orders FOR SELECT USING (
    EXISTS (SELECT 1 FROM admins WHERE id = auth.uid())
  );
CREATE POLICY "Admins can manage orders" 
  ON orders FOR ALL USING (
    EXISTS (SELECT 1 FROM admins WHERE id = auth.uid())
  );

-- Order Items policies: Also restricted to admins
CREATE POLICY "Admins can view all order items" 
  ON order_items FOR SELECT USING (
    EXISTS (SELECT 1 FROM admins WHERE id = auth.uid())
  );
CREATE POLICY "Admins can manage order items" 
  ON order_items FOR ALL USING (
    EXISTS (SELECT 1 FROM admins WHERE id = auth.uid())
  );