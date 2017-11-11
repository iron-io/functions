FROM iron/ruby:dev
RUN apk --no-cache --update add imagemagick

COPY ./ /func
WORKDIR /func

RUN gem install bundler --no-ri --no-rdoc
RUN bundle

ENTRYPOINT bundle exec ruby func.rb
