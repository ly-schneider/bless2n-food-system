-- Allow public to view categories
CREATE POLICY "Public can view categories" 
  ON categories
  FOR SELECT
  USING (TRUE);

-- Only authenticated admins can manage categories
CREATE POLICY "Admins can manage categories" 
  ON categories
  FOR ALL
  USING (EXISTS (SELECT 1 FROM admins WHERE id = auth.uid()));