FROM iron/ruby:dev
RUN apk --no-cache --update add imagemagick

COPY ./ /func
WORKDIR /func

RUN gem install bundler --no-ri --no-rdoc
RUN bundle

ENV AWS_ACCESS_KEY_ID {add your id here}
ENV AWS_SECRET_ACCESS_KEY {add your key here}
ENV AWS_S3_BUCKET iron-functions-image-resize
ENV AWS_REGION us-east-1

ENTRYPOINT bundle exec ruby func.rb
