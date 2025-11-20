#!/bin/bash

# Base Framework Module Generation Test Script
# Tests all field types, relations, and features

echo "========================================"
echo "Base Framework Comprehensive Test"
echo "========================================"
echo ""

# Build the CLI first
echo "Building Base CLI..."
cd ../cmd && go build -o ../base/base . && cd ../base
echo "CLI built successfully!"
echo ""

# 1. Category Module - with translations and hierarchy
echo "1. Generating Category module..."
./base g category \
  name:translation.Field \
  slug:string \
  description:text \
  parent:belongsTo:Category \
  position:int \
  is_active:bool \
  meta_title:string \
  meta_description:text \
  icon:string \
  color:string

# 2. Author Module - with file storage
echo "2. Generating Author module..."
./base g author \
  name:string \
  bio:text \
  email:string \
  avatar:file \
  resume:file \
  website:string \
  twitter:string \
  github:string \
  is_verified:bool \
  rating:float

# 3. Tag Module - for many-to-many relations
echo "3. Generating Tag module..."
./base g tag \
  name:translation.Field \
  slug:string \
  description:translation.Field \
  color:string \
  usage_count:int

# 4. Post Module - main content with all field types
echo "4. Generating Post module..."
./base g post \
  title:translation.Field \
  slug:string \
  content:translation.Field \
  excerpt:translation.Field \
  featured_image:image \
  gallery:file \
  category:belongsTo:Category \
  author:belongsTo:Author \
  status:string \
  published_at:time \
  view_count:int \
  like_count:int \
  is_featured:bool \
  is_premium:bool \
  reading_time:int \
  meta_title:translation.Field \
  meta_description:translation.Field \
  meta_keywords:text

# 5. Comment Module - with relations and moderation
echo "5. Generating Comment module..."
./base g comment \
  post:belongsTo:Post \
  user:belongsTo:profile.User \
  parent:belongsTo:Comment \
  content:text \
  status:string \
  is_approved:bool \
  likes:int \
  dislikes:int \
  ip_address:string \
  user_agent:text

# 6. Media Module - for file management
echo "6. Generating Media module..."
./base g mediaLibrary \
  title:string \
  description:text \
  file:file \
  thumbnail:image \
  mime_type:string \
  size:int \
  width:int \
  height:int \
  duration:int \
  folder:string \
  is_public:bool \
  download_count:int

# 7. Newsletter Module - with email fields
echo "7. Generating Newsletter module..."
./base g newsletter \
  email:string \
  name:string \
  status:string \
  token:string \
  subscribed_at:time \
  unsubscribed_at:time \
  preferences:text \
  source:string

# 8. Event Module - with date/time fields
echo "8. Generating Event module..."
./base g event \
  title:translation.Field \
  description:translation.Field \
  location:string \
  start_date:time \
  end_date:time \
  all_day:bool \
  recurring:bool \
  max_attendees:int \
  current_attendees:int \
  price:float \
  currency:string \
  image:file \
  status:string

# 9. Review Module - with ratings
echo "9. Generating Review module..."
./base g review \
  user:belongsTo:profile.User \
  rating:int \
  title:string \
  content:text \
  pros:text \
  cons:text \
  is_verified:bool \
  helpful_count:int \
  images:file

# 10. Setting Module - for configuration
echo "10. Generating Setting module..."
./base g setting \
  key:string \
  value:text \
  type:string \
  group:string \
  label:translation.Field \
  description:translation.Field \
  is_public:bool \
  is_editable:bool \
  validation_rules:text

echo ""
echo "========================================"
echo "Module generation complete!"
echo "========================================"
echo ""

# Build the modules
echo "Building modules..."
./base build

echo ""
echo "Build complete! You can now test the modules."
echo ""
echo "Suggested tests:"
echo "1. Start server: ./base start"
echo "2. Test each endpoint with all CRUD operations"
echo "3. Test file uploads for avatar, featured_image, etc."
echo "4. Test translations by adding multiple languages"
echo "5. Test relations between posts, categories, authors, etc."