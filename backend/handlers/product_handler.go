package handlers


func CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravni podaci",)
		return
	}

	err = repositories.CreateProduct(&product)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Greska pri kreiranju proizvoda",
		})
		return
	}

	c.JSON(201, product)
}

func GetAllProducts(c *gin.Context) {
	products, err := repositories.GetAllProducts()
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Greska pri ucitavanju proizvoda",
		})
		return
	}

	c.JSON(200, products)
}
